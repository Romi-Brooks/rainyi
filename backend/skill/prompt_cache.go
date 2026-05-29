package skill

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"rain-yi-backend/config"
	"rain-yi-backend/repository"
)

type PromptCache struct {
	mu    sync.RWMutex
	cache map[int64]*cachedPrompt
	repo  *repository.PersonaRepository
}

type cachedPrompt struct {
	Prompt    string
	CreatedAt time.Time
}

func NewPromptCache(repo *repository.PersonaRepository) *PromptCache {
	return &PromptCache{
		cache: make(map[int64]*cachedPrompt),
		repo:  repo,
	}
}

func (pc *PromptCache) Get(personaID int64) (string, bool) {
	pc.mu.RLock()
	cp, ok := pc.cache[personaID]
	pc.mu.RUnlock()
	if ok {
		return cp.Prompt, true
	}

	if config.RDB != nil {
		val, err := config.RDB.Get(config.RedisCtx, fmt.Sprintf("skill:prompt:%d", personaID)).Result()
		if err == nil && val != "" {
			pc.mu.Lock()
			pc.cache[personaID] = &cachedPrompt{Prompt: val, CreatedAt: time.Now()}
			pc.mu.Unlock()
			return val, true
		}
	}

	return "", false
}

func (pc *PromptCache) Set(personaID int64, prompt string) {
	pc.mu.Lock()
	pc.cache[personaID] = &cachedPrompt{Prompt: prompt, CreatedAt: time.Now()}
	pc.mu.Unlock()

	if config.RDB != nil {
		config.RDB.Set(config.RedisCtx, fmt.Sprintf("skill:prompt:%d", personaID), prompt, 0)
	}
}

func (pc *PromptCache) Invalidate(personaID int64) {
	pc.mu.Lock()
	delete(pc.cache, personaID)
	pc.mu.Unlock()

	if config.RDB != nil {
		config.RDB.Del(config.RedisCtx, fmt.Sprintf("skill:prompt:%d", personaID))
	}
}

func (pc *PromptCache) Warmup(personas []struct {
	ID   int64
	Name string
}) {
	for _, p := range personas {
		prompt := pc.buildFromDB(p.ID)
		if prompt != "" {
			pc.Set(p.ID, prompt)
		}
	}
}

func (pc *PromptCache) CompileAndCache(personaID int64) string {
	prompt := pc.buildFromDB(personaID)
	if prompt != "" {
		pc.Set(personaID, prompt)
	}
	return prompt
}

func (pc *PromptCache) buildFromDB(personaID int64) string {
	nodes, err := pc.repo.FindSkillNodesByPersonaID(personaID)
	if err != nil || len(nodes) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, node := range nodes {
		kvs, err := pc.repo.FindSkillKVsBySkillNodeID(node.ID)
		if err == nil && len(kvs) > 0 {
			for _, kv := range kvs {
				builder.WriteString(fmt.Sprintf("# %s\n%s\n\n", kv.Key, kv.Value))
			}
		} else {
			builder.WriteString(node.Content)
			builder.WriteString("\n\n")
		}
	}

	return strings.TrimSpace(builder.String())
}

func (pc *PromptCache) MarshalJSON() ([]byte, error) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	info := make(map[int64]string)
	for id, cp := range pc.cache {
		info[id] = fmt.Sprintf("len=%d age=%s", len(cp.Prompt), time.Since(cp.CreatedAt).Round(time.Second))
	}
	return json.Marshal(info)
}
