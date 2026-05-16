package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type OllamaProvider struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewOllamaProvider(baseURL, model string) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &OllamaProvider{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}
}

func (p *OllamaProvider) Name() string {
	return fmt.Sprintf("Ollama (%s)", p.model)
}

func (p *OllamaProvider) Chat(ctx context.Context, messages []ChatMessage, tools []ToolDef) (*ChatResponse, error) {
	// Ollama /v1 endpoint is OpenAI compatible
	reqBody := openAIChatRequest{
		Model:    p.model,
		Messages: messages,
		Stream:   false,
	}
	
	if len(tools) > 0 {
		for _, t := range tools {
			reqBody.Tools = append(reqBody.Tools, openAITool{
				Type:     "function",
				Function: t,
			})
		}
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error (%d): %s", resp.StatusCode, string(body))
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content   string `json:"content"`
				ToolCalls []struct {
					ID       string `json:"id"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, err
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from ollama")
	}

	res := &ChatResponse{
		Content: chatResp.Choices[0].Message.Content,
	}

	for _, tc := range chatResp.Choices[0].Message.ToolCalls {
		tcNew := ToolCall{
			ID:   tc.ID,
			Type: "function",
		}
		tcNew.Function.Name = tc.Function.Name
		tcNew.Function.Arguments = tc.Function.Arguments
		res.ToolCalls = append(res.ToolCalls, tcNew)
	}

	return res, nil
}

func (p *OllamaProvider) StreamChat(ctx context.Context, messages []ChatMessage, tools []ToolDef) (<-chan StreamChunk, error) {
	// For streaming, we'll use a similar approach to OpenAI
	ch := make(chan StreamChunk)
	
	go func() {
		defer close(ch)
		
		reqBody := openAIChatRequest{
			Model:    p.model,
			Messages: messages,
			Stream:   true,
		}
		
		if len(tools) > 0 {
			for _, t := range tools {
				reqBody.Tools = append(reqBody.Tools, openAITool{
					Type:     "function",
					Function: t,
				})
			}
		}

		jsonData, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err := p.client.Do(req)
		if err != nil {
			ch <- StreamChunk{Error: err}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			ch <- StreamChunk{Error: fmt.Errorf("ollama error (%d): %s", resp.StatusCode, string(body))}
			return
		}

		// Proper SSE decoder loop
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}
			
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				ch <- StreamChunk{Done: true}
				return
			}

			var chunk struct {
				Choices []struct {
					Delta struct {
						Content   string `json:"content"`
						ToolCalls []struct {
							ID       string `json:"id"`
							Function struct {
								Name      string `json:"name"`
								Arguments string `json:"arguments"`
							} `json:"function"`
						} `json:"tool_calls"`
					} `json:"delta"`
					FinishReason *string `json:"finish_reason"`
				} `json:"choices"`
			}

			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}
			
			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta
				chunkRes := StreamChunk{
					Content: delta.Content,
				}
				for _, tc := range delta.ToolCalls {
					tcNew := ToolCall{
						ID:   tc.ID,
						Type: "function",
					}
					tcNew.Function.Name = tc.Function.Name
					tcNew.Function.Arguments = tc.Function.Arguments
					chunkRes.ToolCalls = append(chunkRes.ToolCalls, tcNew)
				}
				if chunk.Choices[0].FinishReason != nil && *chunk.Choices[0].FinishReason != "" {
					chunkRes.Done = true
				}
				ch <- chunkRes
				if chunkRes.Done {
					return
				}
			}
		}
	}()

	return ch, nil
}
