package llm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

var baseURL = "https://ai-gateway.vercel.sh/v1"

var ErrTooManyRequests = errors.New("rate limit exceeded")
var ErrRequest = errors.New("error during completion")

type LLM struct {
	apiKey string
	model  string
}

func NewLLM() *LLM {
	apiKey, exists := os.LookupEnv("VERCEL_AI_GATEWAY_API_KEY")
	if !exists {
		fmt.Fprintln(os.Stderr, "missing VERCEL_AI_GATEWAY_API_KEY environment variable")
		os.Exit(1)
	}

	model, exists := os.LookupEnv("VERCEL_AI_GATEWAY_MODEL")
	if !exists {
		fmt.Fprintln(os.Stderr, "missing VERCEL_AI_GATEWAY_MODEL environment variable, using openai/gpt-6-luna by default")
		model = "openai/gpt-6-luna"
	}

	llm := &LLM{
		apiKey: apiKey,
		model:  model,
	}

	return llm
}

type ChoiceMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CompletionChoice struct {
	Index   int           `json:"index"`
	Message ChoiceMessage `json:"message"`
}

type CompletionResponse struct {
	Id      string             `json:"id"`
	Object  string             `json:"object"`
	Created int                `json:"created"`
	Model   string             `json:"model"`
	Choices []CompletionChoice `json:"choices"`
}

func (l *LLM) Completion(request string) (string, error) {
	client := &http.Client{}
	endpoint := "/chat/completions"

	payload := strings.NewReader(fmt.Sprintf(`{
    "model": "%s",
    "messages": [
      {
        "role": "user",
        "content": "%s"
      }
    ],
    "stream": false
  }`, l.model, request))

	req, err := http.NewRequest("POST", baseURL+endpoint, payload)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", l.apiKey))
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			return "", ErrTooManyRequests
		}
		return "", fmt.Errorf("completion: status %d, %w\n", resp.StatusCode, err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var c CompletionResponse
	if err := json.Unmarshal(body, &c); err != nil {
		return "", fmt.Errorf("decoding completion: %w", err)
	}

	return c.Choices[0].Message.Content, nil
}
