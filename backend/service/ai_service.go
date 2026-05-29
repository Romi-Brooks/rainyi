package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"rain-yi-backend/config"
	"rain-yi-backend/model"
	"rain-yi-backend/skill"
	"rain-yi-backend/utils"
)

type AIService struct {
	skillManager *skill.SkillManager
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Stream      bool          `json:"stream"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

type StreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

const chatCompletionsPath = "/chat/completions"

func (s *AIService) apiURL(baseURL string) string {
	base := strings.TrimRight(baseURL, "/")
	return base + chatCompletionsPath
}

func NewAIService(skillManager *skill.SkillManager) *AIService {
	return &AIService{skillManager: skillManager}
}

func (s *AIService) buildSystemPrompt(conv *model.Conversation) string {
	systemPrompt := s.skillManager.GetSystemPromptByPersona(conv.PersonaID)

	if conv != nil && conv.AINickname != "" {
		systemPrompt = fmt.Sprintf("你的名字是%s。\n%s", conv.AINickname, systemPrompt)
	}

	// TODO: 接入情绪摘要后，从此处注入用户情绪状态
	// emotionSummary, err := s.emotionRepo.FindByConversationID(conv.ID)
	// if err == nil && emotionSummary != nil {
	//     systemPrompt += fmt.Sprintf("\n\n## 用户近期状态\n%s", emotionSummary.Summary)
	// }

	return systemPrompt
}

func (s *AIService) buildMessages(conv *model.Conversation, userMessage string, history []model.Message) []ChatMessage {
	messages := []ChatMessage{
		{Role: "system", Content: s.buildSystemPrompt(conv)},
	}

	for _, msg := range history {
		messages = append(messages, ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: userMessage,
	})

	return messages
}

func (s *AIService) SendMessage(conv *model.Conversation, userMessage string, history []model.Message, onStream func(content string)) (string, error) {
	cfg := config.AppConfig
	if cfg.DeepSeekAPIKey == "" {
		return "", fmt.Errorf("DeepSeek API Key 未配置，请在 .env 文件中设置 DEEPSEEK_API_KEY")
	}

	messages := s.buildMessages(conv, userMessage, history)

	reqBody := ChatRequest{
		Model:       "deepseek-chat",
		Messages:    messages,
		Stream:      true,
		Temperature: 0.8,
		MaxTokens:   2000,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("请求序列化失败: %v", err)
	}

	httpReq, err := http.NewRequest("POST", s.apiURL(cfg.DeepSeekAPIURL), bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+cfg.DeepSeekAPIKey)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("API 请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API 返回错误 [%d]: %s", resp.StatusCode, string(bodyBytes))
	}

	var fullContent strings.Builder
	reader := resp.Body
	buf := make([]byte, 4096)

	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return fullContent.String(), fmt.Errorf("读取流失败: %v", err)
		}

		if n > 0 {
				chunk := string(buf[:n])
				lines := strings.Split(chunk, "\n")

				for _, line := range lines {
					line = strings.TrimSpace(line)
					if !strings.HasPrefix(line, "data: ") {
						continue
					}

					data := strings.TrimPrefix(line, "data: ")
					if data == "[DONE]" {
						continue
					}

					var streamChunk StreamChunk
					if err := json.Unmarshal([]byte(data), &streamChunk); err != nil {
						continue
					}

					if len(streamChunk.Choices) > 0 {
						content := streamChunk.Choices[0].Delta.Content
						if content == "" {
							continue
						}
						fullContent.WriteString(content)

						if onStream != nil {
							onStream(content)
						}
					}
				}
			}

			if err == io.EOF {
				break
			}
		}

	sanitized := utils.SanitizeEmotionalTags(fullContent.String())
	return sanitized, nil
}

func (s *AIService) SendMessageNonStream(conv *model.Conversation, userMessage string, history []model.Message) (string, error) {
	cfg := config.AppConfig
	if cfg.DeepSeekAPIKey == "" {
		return "", fmt.Errorf("DeepSeek API Key 未配置")
	}

	messages := s.buildMessages(conv, userMessage, history)

	reqBody := ChatRequest{
		Model:       "deepseek-chat",
		Messages:    messages,
		Stream:      false,
		Temperature: 0.8,
		MaxTokens:   2000,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("请求序列化失败: %v", err)
	}

	httpReq, err := http.NewRequest("POST", s.apiURL(cfg.DeepSeekAPIURL), bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+cfg.DeepSeekAPIKey)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("API 请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API 返回错误 [%d]: %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if len(chatResp.Choices) > 0 {
		sanitized := utils.SanitizeEmotionalTags(chatResp.Choices[0].Message.Content)
		return sanitized, nil
	}

	return "", fmt.Errorf("AI 返回空响应")
}
