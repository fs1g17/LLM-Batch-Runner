package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

var baseURL = "https://ai-gateway.vercel.sh/v1"

var ErrTooManyRequests = errors.New("rate limit exceeded")
var ErrRequest = errors.New("error during completion")
var ErrMissingEnvVar = errors.New("missing environment variable")

type LLM struct {
	apiKey string
	model  string
}

func NewLLM() (*LLM, error) {
	llm := &LLM{
		apiKey: "",
		model:  "",
	}

	apiKey, exists := os.LookupEnv("VERCEL_AI_GATEWAY_API_KEY")
	if !exists {
		fmt.Fprintln(os.Stderr, "missing VERCEL_AI_GATEWAY_API_KEY environment variable")
		return llm, fmt.Errorf("missing VERCEL_AI_GATEWAY_API_KEY: %w", ErrMissingEnvVar)
	}

	model, exists := os.LookupEnv("VERCEL_AI_GATEWAY_MODEL")
	if !exists {
		fmt.Fprintln(os.Stderr, "missing VERCEL_AI_GATEWAY_MODEL environment variable, using openai/gpt-6-luna by default")
		model = "openai/gpt-6-luna"
	}

	llm.apiKey = apiKey
	llm.model = model

	return llm, nil
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

type CompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CompletionRequest struct {
	Model    string              `json:"model"`
	Messages []CompletionMessage `json:"messages"`
	Stream   bool                `json:"stream"`
}

func (l *LLM) Completion(ctx context.Context, request string) (string, error) {
	client := &http.Client{}
	endpoint := "/chat/completions"

	payload, err := json.Marshal(CompletionRequest{
		Model: l.model,
		Messages: []CompletionMessage{
			{
				Role:    "user",
				Content: request,
			},
		},
		Stream: false,
	})

	if err != nil {
		return "", fmt.Errorf("marshall completion request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+endpoint, bytes.NewReader(payload))
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
