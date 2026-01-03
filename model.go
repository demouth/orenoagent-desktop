package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/demouth/orenoagent-go"
	"github.com/demouth/orenoagent-go/provider/openai"
	openaiSDK "github.com/openai/openai-go/v3"

	"github.com/tectiv3/websearch"
	"github.com/tectiv3/websearch/provider"
)

type Model struct {
	interaction []interactionResult

	mu      sync.RWMutex
	running bool
}

type interactionResult interface {
	Type() string
}

var agent *orenoagent.Agent

type askInteraction struct {
	prompt string
}

func (*askInteraction) Type() string {
	return "askInteraction"
}
func (a *askInteraction) String() string {
	return a.prompt
}

func init() {

	var tools = []orenoagent.Tool{
		{
			Name:        "currentTime",
			Description: "Get the current date and time with timezone in a human-readable format.",
			Function: func(_ string) string {
				return time.Now().Format(time.RFC3339)
			},
		},
		{
			// NOTE: This is a sample function. Do not use it in production environments.

			Name:        "webSearch",
			Description: "Get the current date and time with timezone in a human-readable format.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"keyword": map[string]string{
						"type":        "string",
						"description": "web search keyword.",
					},
				},
				"required": []string{"keyword"},
			},
			Function: func(args string) string {
				var param struct {
					Keyword string
				}
				err := json.Unmarshal([]byte(args), &param)
				if err != nil {
					return fmt.Sprintf("%v", err)
				}

				type result struct {
					Title   string
					Link    string
					Snippet string
				}
				results := []result{}
				web := websearch.New(provider.NewUnofficialDuckDuckGo())
				res, err := web.Search(param.Keyword, 10)
				if err != nil {
					return fmt.Sprintf("%v", err)
				}
				for _, ddgor := range res {
					r := result{
						Title:   ddgor.Title,
						Link:    ddgor.Link.String(),
						Snippet: ddgor.Description,
					}
					results = append(results, r)
				}
				v, _ := json.Marshal(results)

				return string(v)
			},
		},
		{
			Name:        "WebReader",
			Description: "Reads and returns the content from the specified URL",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"url": map[string]string{
						"type":        "string",
						"description": "URL of the page to retrieve",
					},
				},
				"required": []string{"url"},
			},
			Function: func(args string) string {
				var param struct {
					Url string
				}
				err := json.Unmarshal([]byte(args), &param)
				if err != nil {
					return fmt.Sprintf("%v", err)
				}

				req, _ := http.NewRequest("GET", param.Url, nil)
				req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					return fmt.Sprintf("%v", err)
				}
				defer resp.Body.Close()
				bodyBytes, err := io.ReadAll(resp.Body)
				if err != nil {
					return fmt.Sprintf("%v", err)
				}

				return string(bodyBytes)
			},
		},
	}

	client := openaiSDK.NewClient()
	provider := openai.NewProvider(client)
	agent = orenoagent.NewAgent(
		provider,
		orenoagent.WithTools(tools),
		orenoagent.WithReasoningSummary("detailed"),
		orenoagent.WithReasoningEffort("low"),
		orenoagent.WithModel(openaiSDK.ChatModelGPT5Nano),
	)

}

func (m *Model) Running() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

func (m *Model) CanAsk(prompt string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.running {
		return false
	}
	return strings.TrimSpace(prompt) != ""
}

func (m *Model) TryAsk(prompt string) bool {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return false
	}

	ctx := context.Background() // TODO: Use a proper context.
	subscriber, _ := agent.Ask(ctx, prompt)

	m.mu.RLock()
	if m.running {
		m.mu.RUnlock()
		return false
	}
	m.mu.RUnlock()

	m.mu.Lock()
	m.running = true
	m.interaction = append(m.interaction, &askInteraction{prompt: prompt})
	m.mu.Unlock()

	println("Started processing prompt:", prompt)

	go func() {
		for result := range subscriber.Subscribe() {
			println(result.Type())
			switch r := result.(type) {
			case *orenoagent.ErrorResult:
				println(r.Error())
				return
			case *orenoagent.MessageDeltaResult:
				m.mu.Lock()
				m.interaction = append(m.interaction, r)
				m.mu.Unlock()
			case *orenoagent.ReasoningDeltaResult:
				m.mu.Lock()
				m.interaction = append(m.interaction, r)
				m.mu.Unlock()
			case *orenoagent.FunctionCallResult:
				m.mu.Lock()
				m.interaction = append(m.interaction, r)
				m.mu.Unlock()
			}
		}

		m.mu.Lock()
		m.running = false
		m.mu.Unlock()
	}()

	return true
}

func (m *Model) InteractionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.interaction)
}

func (m *Model) InteractionByIndex(i int) interactionResult {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.interaction[i]
}
