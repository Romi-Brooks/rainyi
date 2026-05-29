package controller

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"rain-yi-backend/config"
	"rain-yi-backend/model"
	"rain-yi-backend/repository"
	"rain-yi-backend/service"
	"rain-yi-backend/skill"
	"rain-yi-backend/utils"

	"github.com/gin-gonic/gin"
)

type PersonaController struct {
	personaRepo *repository.PersonaRepository
	convRepo    *repository.ConversationRepository
	promptCache *skill.PromptCache
	storage     service.FileStorage
	fileRepo    *repository.FileRepository
}

func NewPersonaController(personaRepo *repository.PersonaRepository, convRepo *repository.ConversationRepository, promptCache *skill.PromptCache, storage service.FileStorage, fileRepo *repository.FileRepository) *PersonaController {
	return &PersonaController{
		personaRepo: personaRepo,
		convRepo:    convRepo,
		promptCache: promptCache,
		storage:     storage,
		fileRepo:    fileRepo,
	}
}

func (ctl *PersonaController) GetPersonas(c *gin.Context) {
	userID := c.GetInt64("user_id")

	personas, err := ctl.personaRepo.FindAllVisible(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取人格列表失败"})
		return
	}

	type personaWithCount struct {
		model.Persona
		SkillNodeCount int64 `json:"skill_node_count"`
		IsBuiltIn      bool  `json:"is_built_in"`
	}

	result := make([]personaWithCount, 0, len(personas))
	for _, p := range personas {
		count, _ := ctl.personaRepo.CountSkillNodesByPersonaID(p.ID)
		result = append(result, personaWithCount{
			Persona:        p,
			SkillNodeCount: count,
			IsBuiltIn:      p.UserID == 0,
		})
	}

	c.JSON(http.StatusOK, gin.H{"personas": result})
}

func (ctl *PersonaController) GetPersona(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "人格ID无效"})
		return
	}

	persona, err := ctl.personaRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "人格不存在"})
		return
	}

	nodes, err := ctl.personaRepo.FindSkillNodesByPersonaID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取技能节点失败"})
		return
	}

	type skillNodeWithKVs struct {
		model.SkillNode
		KVs []model.SkillKV `json:"kvs"`
	}

	resultNodes := make([]skillNodeWithKVs, 0, len(nodes))
	for _, node := range nodes {
		kvs, _ := ctl.personaRepo.FindSkillKVsBySkillNodeID(node.ID)
		resultNodes = append(resultNodes, skillNodeWithKVs{
			SkillNode: node,
			KVs:       kvs,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"persona":      persona,
		"skill_nodes":  resultNodes,
	})
}

func (ctl *PersonaController) CreatePersona(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "人格名称不能为空"})
		return
	}

	persona := &model.Persona{
		UserID:      userID,
		Name:        utils.SanitizeInput(req.Name),
		Description: utils.SanitizeInput(req.Description),
	}
	if err := ctl.personaRepo.Create(persona); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建人格失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "创建成功",
		"persona": persona,
	})
}

func (ctl *PersonaController) UpdatePersona(c *gin.Context) {
	userID := c.GetInt64("user_id")
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "人格ID无效"})
		return
	}

	persona, err := ctl.personaRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "人格不存在"})
		return
	}

	if persona.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权修改该人格"})
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if req.Name != nil {
		persona.Name = utils.SanitizeInput(*req.Name)
	}
	if req.Description != nil {
		persona.Description = utils.SanitizeInput(*req.Description)
	}

	if err := ctl.personaRepo.Update(persona); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"persona": persona,
	})
}

func (ctl *PersonaController) DeletePersona(c *gin.Context) {
	userID := c.GetInt64("user_id")
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "人格ID无效"})
		return
	}

	persona, err := ctl.personaRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "人格不存在"})
		return
	}

	if persona.UserID != userID && persona.UserID != 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除该人格"})
		return
	}

	if persona.UserID == 0 && userID != 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除内置人格"})
		return
	}

	ctl.personaRepo.Delete(id)

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func (ctl *PersonaController) UploadSkillFile(c *gin.Context) {
	userID := c.GetInt64("user_id")
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "人格ID无效"})
		return
	}

	persona, err := ctl.personaRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "人格不存在"})
		return
	}

	if persona.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作该人格"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供技能文件"})
		return
	}

	files := form.File["file"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请至少选择一个 .md 文件"})
		return
	}

	var uploaded []string
	var errors []string

	for _, header := range files {
		if !strings.HasSuffix(header.Filename, ".md") {
			errors = append(errors, fmt.Sprintf("跳过非 .md 文件: %s", header.Filename))
			continue
		}

		file, err := header.Open()
		if err != nil {
			errors = append(errors, fmt.Sprintf("打开文件失败 %s: %v", header.Filename, err))
			continue
		}

		content, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			errors = append(errors, fmt.Sprintf("读取文件失败 %s: %v", header.Filename, err))
			continue
		}

		parsed, err := skill.ParseSkillContent(header.Filename, string(content))
		if err != nil {
			errors = append(errors, fmt.Sprintf("解析文件失败 %s: %v", header.Filename, err))
			continue
		}

		var storagePath string
		var fileRecordID int64
		if ctl.storage != nil {
			record, err := service.SaveUploadedFile(ctl.storage, "skill_raw", userID, "skill_node", 0, header)
			if err == nil {
				storagePath = record.StoragePath
				fileRecordID = record.ID
			}
		}

		nextPriority := 0
		nodes, _ := ctl.personaRepo.FindSkillNodesByPersonaID(id)
		if len(nodes) > 0 {
			nextPriority = nodes[len(nodes)-1].Priority + 1
		}

		node := &model.SkillNode{
			PersonaID:   id,
			Name:        parsed.Meta.Name,
			Description: parsed.Meta.Description,
			FileName:    header.Filename,
			Content:     string(content),
			Source:      "db",
			Priority:    nextPriority,
		}
		if storagePath != "" {
			node.Source = "fs"
			node.StoragePath = storagePath
		}
		if err := ctl.personaRepo.CreateSkillNode(node); err != nil {
			errors = append(errors, fmt.Sprintf("保存节点失败 %s: %v", header.Filename, err))
			continue
		}

		if fileRecordID > 0 && ctl.fileRepo != nil {
			rec, _ := ctl.fileRepo.FindByID(fileRecordID)
			if rec != nil {
				rec.ReferenceID = node.ID
				rec.ReferenceType = "skill_node"
				ctl.fileRepo.Update(rec)
			}
		}

		if len(parsed.KVList) > 0 {
			kvs := make([]model.SkillKV, 0, len(parsed.KVList))
			for ki, kv := range parsed.KVList {
				kvs = append(kvs, model.SkillKV{
					SkillNodeID: node.ID,
					Key:         kv.Key,
					Value:       kv.Value,
					SortOrder:   ki,
				})
			}
			if err := ctl.personaRepo.CreateSkillKVsBatch(kvs); err != nil {
				errors = append(errors, fmt.Sprintf("保存键值对失败 %s: %v", header.Filename, err))
				continue
			}
		}

		uploaded = append(uploaded, header.Filename)
	}

	if ctl.promptCache != nil && len(uploaded) > 0 {
		ctl.promptCache.Invalidate(id)
	}

	response := gin.H{"message": fmt.Sprintf("成功上传 %d 个文件", len(uploaded))}
	if len(uploaded) > 0 {
		response["uploaded"] = uploaded
	}
	if len(errors) > 0 {
		response["errors"] = errors
	}

	c.JSON(http.StatusOK, response)
}

func (ctl *PersonaController) DeleteSkillFile(c *gin.Context) {
	userID := c.GetInt64("user_id")
	personaIDStr := c.Param("id")
	fileIDStr := c.Param("fileId")

	personaID, err := strconv.ParseInt(personaIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "人格ID无效"})
		return
	}

	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "技能节点ID无效"})
		return
	}

	persona, err := ctl.personaRepo.FindByID(personaID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "人格不存在"})
		return
	}

	if persona.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作该人格"})
		return
	}

	node, err := ctl.personaRepo.FindSkillNodeByID(fileID)
	if err != nil || node.PersonaID != personaID {
		c.JSON(http.StatusNotFound, gin.H{"error": "技能节点不存在"})
		return
	}

	if node.Source == "fs" && node.StoragePath != "" && ctl.storage != nil {
		ctl.storage.Delete(&model.FileRecord{StoragePath: node.StoragePath})
	}

	ctl.personaRepo.DeleteSkillNode(fileID)

	if ctl.promptCache != nil {
		ctl.promptCache.Invalidate(personaID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func (ctl *PersonaController) SetConversationPersona(c *gin.Context) {
	userID := c.GetInt64("user_id")
	convIDStr := c.Param("id")
	convID, err := strconv.ParseInt(convIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话ID无效"})
		return
	}

	conv, err := ctl.convRepo.FindByID(convID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
		return
	}

	if conv.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作该会话"})
		return
	}

	var req struct {
		PersonaID *int64 `json:"persona_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if req.PersonaID != nil {
		if _, err := ctl.personaRepo.FindByID(*req.PersonaID); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "人格不存在"})
			return
		}
	}

	if err := ctl.personaRepo.SetConversationPersona(convID, req.PersonaID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "设置失败"})
		return
	}

	conv.PersonaID = req.PersonaID

	c.JSON(http.StatusOK, gin.H{
		"message":      "设置成功",
		"conversation": conv,
	})
}

func (ctl *PersonaController) LoadFromDirectory(c *gin.Context) {
	userID := c.GetInt64("user_id")

	skillsDir := config.AppConfig.SkillsDir
	absDir, err := filepath.Abs(skillsDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "技能目录解析失败"})
		return
	}
	absDir = filepath.Clean(absDir)

	entries, err := os.ReadDir(absDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("读取技能目录失败: %v", err)})
		return
	}

	loadedCount := 0
	loadedPersonas := make(map[string]int64)

	for _, entry := range entries {
		if entry.IsDir() {
			dirPath := filepath.Join(absDir, entry.Name())
			dirName := entry.Name()

			dirEntries, _ := os.ReadDir(dirPath)
			var validMdFiles []string
			for _, de := range dirEntries {
				if !de.IsDir() && strings.HasSuffix(de.Name(), ".md") {
					validMdFiles = append(validMdFiles, de.Name())
				}
			}
			if len(validMdFiles) == 0 {
				continue
			}

			existing, _ := ctl.personaRepo.FindByName(dirName)
			if existing != nil {
				if existing.UserID != 0 && existing.UserID != userID {
					continue
				}
				ctl.personaRepo.DeleteSkillNodesByPersonaID(existing.ID)
				ctl.personaRepo.Delete(existing.ID)
			}

			persona := &model.Persona{
				UserID:      0,
				Name:        dirName,
				Description: fmt.Sprintf("包含 %d 个技能文件的人格", len(validMdFiles)),
			}
			if err := ctl.personaRepo.Create(persona); err != nil {
				continue
			}

			for i, fn := range validMdFiles {
				filePath := filepath.Join(dirPath, fn)
				rawContent, err := os.ReadFile(filePath)
				if err != nil {
					continue
				}

				parsed, err := skill.ParseSkillContent(fn, string(rawContent))
				if err != nil {
					continue
				}

				node := &model.SkillNode{
					PersonaID:   persona.ID,
					Name:        parsed.Meta.Name,
					Description: parsed.Meta.Description,
					FileName:    fn,
					Content:     string(rawContent),
					Priority:    i,
				}
				if err := ctl.personaRepo.CreateSkillNode(node); err != nil {
					continue
				}

				if len(parsed.KVList) > 0 {
					kvs := make([]model.SkillKV, 0, len(parsed.KVList))
					for ki, kv := range parsed.KVList {
						kvs = append(kvs, model.SkillKV{
							SkillNodeID: node.ID,
							Key:         kv.Key,
							Value:       kv.Value,
							SortOrder:   ki,
						})
					}
					ctl.personaRepo.CreateSkillKVsBatch(kvs)
				}
			}
			loadedPersonas[dirName] = persona.ID
			loadedCount++

		} else if strings.HasSuffix(entry.Name(), ".md") {
			filePath := filepath.Join(absDir, entry.Name())
			rawContent, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}

			parsed, err := skill.ParseSkillContent(entry.Name(), string(rawContent))
			if err != nil {
				continue
			}

			personaName := parsed.Meta.Name
			existing, _ := ctl.personaRepo.FindByName(personaName)
			if existing != nil {
				if existing.UserID != 0 && existing.UserID != userID {
					continue
				}
			} else {
				persona := &model.Persona{
					UserID:      0,
					Name:        personaName,
					Description: parsed.Meta.Description,
				}
				if err := ctl.personaRepo.Create(persona); err != nil {
					continue
				}
				existing = persona
			}

			node := &model.SkillNode{
				PersonaID:   existing.ID,
				Name:        parsed.Meta.Name,
				Description: parsed.Meta.Description,
				FileName:    entry.Name(),
				Content:     string(rawContent),
				Priority:    0,
			}
			if err := ctl.personaRepo.CreateSkillNode(node); err != nil {
				continue
			}

			if len(parsed.KVList) > 0 {
				kvs := make([]model.SkillKV, 0, len(parsed.KVList))
				for ki, kv := range parsed.KVList {
					kvs = append(kvs, model.SkillKV{
						SkillNodeID: node.ID,
						Key:         kv.Key,
						Value:       kv.Value,
						SortOrder:   ki,
					})
				}
				ctl.personaRepo.CreateSkillKVsBatch(kvs)
			}
		}
		loadedCount++
	}

	if ctl.promptCache != nil {
		for _, personaID := range loadedPersonas {
			ctl.promptCache.Invalidate(personaID)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("成功加载 %d 个人格", loadedCount),
		"count":   loadedCount,
	})
}

func (ctl *PersonaController) GetConversationPersona(c *gin.Context) {
	userID := c.GetInt64("user_id")
	convIDStr := c.Param("id")
	convID, err := strconv.ParseInt(convIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话ID无效"})
		return
	}

	conv, err := ctl.convRepo.FindByID(convID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
		return
	}

	if conv.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该会话"})
		return
	}

	if conv.PersonaID == nil {
		c.JSON(http.StatusOK, gin.H{"persona": nil})
		return
	}

	persona, err := ctl.personaRepo.FindByID(*conv.PersonaID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"persona": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"persona": persona})
}
