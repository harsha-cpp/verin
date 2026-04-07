package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type aiAnalysis struct {
	Summary   string   `json:"summary"`
	Findings  []string `json:"findings"`
	Contacts  []string `json:"contacts"`
	Sensitive []string `json:"sensitive"`
}

type aiCache struct {
	mu    sync.RWMutex
	items map[string]aiFlight
}

type aiFlight struct {
	result aiAnalysis
	at     time.Time
}

var aiStore = &aiCache{items: make(map[string]aiFlight)}

func (c *aiCache) get(key string) (aiAnalysis, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	f, ok := c.items[key]
	if !ok || time.Since(f.at) > 24*time.Hour {
		return aiAnalysis{}, false
	}
	return f.result, true
}

func (c *aiCache) set(key string, result aiAnalysis) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = aiFlight{result: result, at: time.Now()}
}

const aiPrompt = `Analyze this document text. Reply using ONLY this exact format, no markdown, no emojis:

SUMMARY
[1-2 plain sentences describing what this document is about]

FINDINGS
[key fact or point from the document]
[key fact or point from the document]
[key fact or point from the document]

CONTACTS
[every email address, phone number, or URL found — one per line. Write None if not found]

SENSITIVE
[any account numbers, IDs, passwords, license keys, SSNs, or financial data — one per line. Write None if not found]

Document text:
`

const openRouterModel = "arcee-ai/trinity-large-preview:free"
const openRouterURL = "https://openrouter.ai/api/v1/chat/completions"

func callGemini(ctx context.Context, apiKey, versionID, content string) (aiAnalysis, error) {
	if result, ok := aiStore.get(versionID); ok {
		return result, nil
	}

	if strings.TrimSpace(content) == "" {
		return aiAnalysis{}, fmt.Errorf("empty content")
	}

	truncated := content
	if len(truncated) > 8000 {
		truncated = truncated[:8000]
	}

	body, _ := json.Marshal(map[string]any{
		"model": openRouterModel,
		"messages": []map[string]any{
			{"role": "user", "content": aiPrompt + truncated},
		},
		"max_tokens":  600,
		"temperature": 0.1,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openRouterURL, bytes.NewReader(body))
	if err != nil {
		return aiAnalysis{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("HTTP-Referer", "https://verin.app")
	req.Header.Set("X-Title", "Verin DMS")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return aiAnalysis{}, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return aiAnalysis{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return aiAnalysis{}, fmt.Errorf("openrouter %d: %s", resp.StatusCode, string(raw))
	}

	var payload struct {
		Choices []struct {
			Message struct {
				Content   *string `json:"content"`
				Reasoning string  `json:"reasoning"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return aiAnalysis{}, err
	}
	if payload.Error != nil {
		return aiAnalysis{}, fmt.Errorf("openrouter error: %s", payload.Error.Message)
	}
	if len(payload.Choices) == 0 {
		return aiAnalysis{}, fmt.Errorf("empty openrouter response")
	}

	text := ""
	msg := payload.Choices[0].Message
	if msg.Content != nil && strings.TrimSpace(*msg.Content) != "" {
		text = *msg.Content
	} else if msg.Reasoning != "" {
		text = extractStructuredFromReasoning(msg.Reasoning)
	}

	if text == "" {
		return aiAnalysis{}, fmt.Errorf("no usable content in response")
	}

	result := parseAIText(text)
	aiStore.set(versionID, result)
	return result, nil
}

func extractStructuredFromReasoning(reasoning string) string {
	idx := strings.Index(reasoning, "SUMMARY")
	if idx == -1 {
		return ""
	}
	return reasoning[idx:]
}

func parseAIText(text string) aiAnalysis {
	var result aiAnalysis
	result.Findings = []string{}
	result.Contacts = []string{}
	result.Sensitive = []string{}

	sections := map[string]*[]string{
		"FINDINGS":  &result.Findings,
		"CONTACTS":  &result.Contacts,
		"SENSITIVE": &result.Sensitive,
	}

	var currentSection string
	var summaryLines []string

	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		// Strip markdown bold: **SUMMARY** → SUMMARY
		stripped := strings.TrimPrefix(strings.TrimSuffix(line, "**"), "**")
		upper := strings.ToUpper(stripped)

		if upper == "SUMMARY" {
			currentSection = "SUMMARY"
			continue
		}
		if _, ok := sections[upper]; ok {
			currentSection = upper
			continue
		}

		switch currentSection {
		case "SUMMARY":
			summaryLines = append(summaryLines, stripped)
		default:
			if target, ok := sections[currentSection]; ok {
				clean := strings.TrimLeft(stripped, "-•*[] ")
				// strip trailing brackets like [key fact] → key fact
				clean = strings.TrimRight(clean, "]")
				if clean != "" && strings.ToLower(clean) != "none" {
					*target = append(*target, clean)
				}
			}
		}
	}

	result.Summary = strings.Join(summaryLines, " ")
	return result
}
