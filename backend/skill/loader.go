package skill

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"rain-yi-backend/config"
	"rain-yi-backend/model"
	"rain-yi-backend/repository"

	"gopkg.in/yaml.v3"
)

type SkillMeta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type ParsedSkill struct {
	Meta     SkillMeta
	FileName string
	KVList   []ParsedKV
}

type ParsedKV struct {
	Key   string
	Value string
}

type SkillManager struct {
	mu          sync.RWMutex
	repo        *repository.PersonaRepository
	promptCache *PromptCache
}

func NewSkillManager(repo *repository.PersonaRepository) *SkillManager {
	return &SkillManager{
		repo:        repo,
		promptCache: NewPromptCache(repo),
	}
}

func parseSkillFile(filePath string) (*ParsedSkill, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(lines) < 2 || lines[0] != "---" {
		return nil, fmt.Errorf("no valid frontmatter found")
	}

	endIndex := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			endIndex = i
			break
		}
	}
	if endIndex == -1 {
		return nil, fmt.Errorf("no closing frontmatter delimiter found")
	}

	metaYaml := strings.Join(lines[1:endIndex], "\n")
	var meta SkillMeta
	if err := yaml.Unmarshal([]byte(metaYaml), &meta); err != nil {
		return nil, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	if meta.Name == "" {
		meta.Name = strings.TrimSuffix(filepath.Base(filePath), ".md")
	}

	bodyLines := lines[endIndex+1:]
	kvList := parseKVFromBody(bodyLines)

	return &ParsedSkill{
		Meta:     meta,
		FileName: filepath.Base(filePath),
		KVList:   kvList,
	}, nil
}

var headingRegex = regexp.MustCompile(`^#{1,6}\s+(.+)$`)

func ParseSkillContent(fileName string, content string) (*ParsedSkill, error) {
	lines := strings.Split(content, "\n")

	var meta SkillMeta
	var bodyLines []string

	if len(lines) >= 2 && strings.TrimSpace(lines[0]) == "---" {
		endIndex := -1
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				endIndex = i
				break
			}
		}

		if endIndex != -1 {
			metaYaml := strings.Join(lines[1:endIndex], "\n")
			if err := yaml.Unmarshal([]byte(metaYaml), &meta); err == nil {
				bodyLines = lines[endIndex+1:]
			} else {
				bodyLines = lines
			}
		} else {
			bodyLines = lines
		}
	} else {
		bodyLines = lines
	}

	if meta.Name == "" {
		meta.Name = strings.TrimSuffix(fileName, ".md")
	}

	kvList := parseKVFromBody(bodyLines)

	return &ParsedSkill{
		Meta:     meta,
		FileName: fileName,
		KVList:   kvList,
	}, nil
}

func parseKVFromBody(lines []string) []ParsedKV {
	var kvs []ParsedKV
	var currentKey string
	var currentValueLines []string

	flush := func() {
		if currentKey != "" {
			value := strings.TrimSpace(strings.Join(currentValueLines, "\n"))
			if value != "" {
				kvs = append(kvs, ParsedKV{Key: currentKey, Value: value})
			}
		}
	}

	for _, line := range lines {
		matches := headingRegex.FindStringSubmatch(line)
		if matches != nil {
			flush()
			currentKey = strings.TrimSpace(matches[1])
			currentValueLines = nil
		} else if currentKey != "" {
			currentValueLines = append(currentValueLines, line)
		}
	}
	flush()

	if len(kvs) == 0 {
		content := strings.TrimSpace(strings.Join(lines, "\n"))
		if content != "" {
			kvs = append(kvs, ParsedKV{Key: "content", Value: content})
		}
	}

	return kvs
}

func (m *SkillManager) LoadSkills() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	skillsDir := config.AppConfig.SkillsDir
	absDir, err := filepath.Abs(skillsDir)
	if err != nil {
		return fmt.Errorf("failed to resolve skills directory: %w", err)
	}
	absDir = filepath.Clean(absDir)

	entries, err := os.ReadDir(absDir)
	if err != nil {
		return fmt.Errorf("failed to read skills directory %s: %w", absDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirPath := filepath.Join(absDir, entry.Name())
			if err := m.loadPersonaFromDir(dirPath, entry.Name()); err != nil {
				continue
			}
		} else if strings.HasSuffix(entry.Name(), ".md") {
			filePath := filepath.Join(absDir, entry.Name())
			filePath = filepath.Clean(filePath)

			if !strings.HasPrefix(filePath, absDir) {
				continue
			}

			if err := m.loadSingleMDFile(filePath); err != nil {
				continue
			}
		}
	}

	personas, _ := m.repo.FindAllVisible(0)
	for _, p := range personas {
		m.promptCache.CompileAndCache(p.ID)
	}

	return nil
}

func (m *SkillManager) loadSingleMDFile(filePath string) error {
	parsed, err := parseSkillFile(filePath)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	personaName := parsed.Meta.Name
	existing, _ := m.repo.FindByName(personaName)
	if existing == nil {
		persona := &model.Persona{
			UserID:      0,
			Name:        personaName,
			Description: parsed.Meta.Description,
		}
		if err := m.repo.Create(persona); err != nil {
			return fmt.Errorf("failed to create persona %s: %w", personaName, err)
		}
		existing = persona
	}

	if err := m.createSkillNodeWithKVs(existing.ID, parsed, string(content), 0); err != nil {
		return fmt.Errorf("failed to save skill %s: %w", parsed.FileName, err)
	}

	m.promptCache.Invalidate(existing.ID)

	return nil
}

func (m *SkillManager) loadPersonaFromDir(dirPath string, dirName string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	var parsedSkills []*ParsedSkill
	var rawContents []string

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		filePath = filepath.Clean(filePath)

		parsed, err := parseSkillFile(filePath)
		if err != nil {
			continue
		}

		rawContent, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		parsedSkills = append(parsedSkills, parsed)
		rawContents = append(rawContents, string(rawContent))
	}

	if len(parsedSkills) == 0 {
		return fmt.Errorf("no valid skill files in directory %s", dirName)
	}

	existing, _ := m.repo.FindByName(dirName)
	if existing != nil {
		m.repo.DeleteSkillNodesByPersonaID(existing.ID)
		m.repo.Delete(existing.ID)
	}

	persona := &model.Persona{
		UserID:      0,
		Name:        dirName,
		Description: fmt.Sprintf("包含 %d 个技能文件的人格", len(parsedSkills)),
	}
	if err := m.repo.Create(persona); err != nil {
		return fmt.Errorf("failed to create persona from dir %s: %w", dirName, err)
	}

	for i, ps := range parsedSkills {
		if err := m.createSkillNodeWithKVs(persona.ID, ps, rawContents[i], i); err != nil {
			return err
		}
	}

	m.promptCache.Invalidate(persona.ID)

	return nil
}

func (m *SkillManager) createSkillNodeWithKVs(personaID int64, parsed *ParsedSkill, content string, priority int) error {
	node := &model.SkillNode{
		PersonaID:   personaID,
		Name:        parsed.Meta.Name,
		Description: parsed.Meta.Description,
		FileName:    parsed.FileName,
		Content:     content,
		Priority:    priority,
	}
	if err := m.repo.CreateSkillNode(node); err != nil {
		return fmt.Errorf("failed to create skill node: %w", err)
	}

	if len(parsed.KVList) > 0 {
		kvs := make([]model.SkillKV, 0, len(parsed.KVList))
		for i, kv := range parsed.KVList {
			kvs = append(kvs, model.SkillKV{
				SkillNodeID: node.ID,
				Key:         kv.Key,
				Value:       kv.Value,
				SortOrder:   i,
			})
		}
		if err := m.repo.CreateSkillKVsBatch(kvs); err != nil {
			return fmt.Errorf("failed to create skill KVs: %w", err)
		}
	}

	return nil
}

func (m *SkillManager) GetSystemPromptByPersona(personaID *int64) string {
	if personaID == nil {
		return "你是一个温柔、耐心、治愈的情感陪伴助手，像一个温暖的朋友。"
	}

	if cached, ok := m.promptCache.Get(*personaID); ok {
		return cached
	}

	prompt := m.promptCache.CompileAndCache(*personaID)
	if prompt == "" {
		return "你是一个温柔、耐心、治愈的情感陪伴助手，像一个温暖的朋友。"
	}

	return prompt
}

func (m *SkillManager) GetPersonaNames() ([]string, error) {
	personas, err := m.repo.FindAllVisible(0)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(personas))
	for _, p := range personas {
		names = append(names, p.Name)
	}
	return names, nil
}

func (m *SkillManager) Refresh() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.repo.DeleteAllByUserID(0); err != nil {
		return fmt.Errorf("failed to clear built-in personas: %w", err)
	}

	return nil
}

func (m *SkillManager) PromptCache() *PromptCache {
	return m.promptCache
}
