// Package ai isolates the AI Legal Assistant's provider integration. Groq is
// the only provider used; nothing outside this package should know about
// Groq-specific request/response shapes, errors, or credentials.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const groqChatCompletionsURL = "https://api.groq.com/openai/v1/chat/completions"

// genericFailureMessage is the only error text ever returned to callers.
// Real provider errors (rate limits, auth failures, outages) are logged
// server-side only and never surface to the Flutter app.
const genericFailureMessage = "AI service is temporarily unavailable. Please try again later."

// systemPrompt is intentionally short to minimize token usage on every call.
const systemPrompt = "You are an AI Legal Assistant for Indian lawyers. Give accurate, concise legal information. " +
	"Keep responses under 250 words unless the user explicitly asks for detailed explanations."

const (
	defaultModel   = "llama-3.1-8b-instant"
	visionModel    = "meta-llama/llama-4-scout-17b-16e-instruct"
	requestTimeout = 20 * time.Second
	maxTokens      = 350
	ocrMaxTokens   = 1024
	temperature    = 0.2
)

// ocrSystemPrompt keeps the OCR call cheap and focused: extract text only,
// no commentary, no formatting narration.
const ocrSystemPrompt = "Extract all readable text from this image exactly as it appears. " +
	"Return only the extracted text, with no commentary or explanation."

var httpClient = &http.Client{Timeout: requestTimeout}

type groqChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqChatRequest struct {
	Model       string            `json:"model"`
	Messages    []groqChatMessage `json:"messages"`
	Temperature float64           `json:"temperature"`
	MaxTokens   int               `json:"max_tokens"`
}

type groqChatResponse struct {
	Choices []struct {
		Message groqChatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func apiKey() string {
	return os.Getenv("GROQ_API_KEY")
}

func model() string {
	if m := os.Getenv("GROQ_MODEL"); m != "" {
		return m
	}
	return defaultModel
}

// send posts a pre-marshaled Groq chat-completions request and returns the
// first choice's text. Every failure path collapses to the generic,
// user-safe error message — never the underlying provider error or key.
func send(ctx context.Context, payload []byte) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, groqChatCompletionsURL, bytes.NewReader(payload))
	if err != nil {
		log.Printf("[ai] request build failed: %v", err)
		return "", errors.New(genericFailureMessage)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey())

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("[ai] groq request failed: %v", err)
		return "", errors.New(genericFailureMessage)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ai] failed to read groq response body: %v", err)
		return "", errors.New(genericFailureMessage)
	}

	var parsed groqChatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		log.Printf("[ai] failed to parse groq response (status %d): %v; body=%s", resp.StatusCode, err, body)
		return "", errors.New(genericFailureMessage)
	}

	if resp.StatusCode != http.StatusOK || parsed.Error != nil {
		providerMsg := ""
		if parsed.Error != nil {
			providerMsg = parsed.Error.Message
		}
		log.Printf("[ai] groq returned an error: status=%d message=%q", resp.StatusCode, providerMsg)
		return "", errors.New(genericFailureMessage)
	}

	if len(parsed.Choices) == 0 || parsed.Choices[0].Message.Content == "" {
		log.Printf("[ai] groq returned no choices/content (status %d): body=%s", resp.StatusCode, body)
		return "", errors.New(genericFailureMessage)
	}

	return parsed.Choices[0].Message.Content, nil
}

// GenerateResponse sends a single prompt to Groq using a fixed, concise
// system prompt and low-token generation settings. On any failure (missing
// key, network error, non-200 response, malformed response) it returns the
// generic, user-safe error message — never the underlying provider error.
func GenerateResponse(ctx context.Context, prompt string) (string, error) {
	if apiKey() == "" {
		log.Printf("[ai] GROQ_API_KEY is not set in this process's environment")
		return "", errors.New(genericFailureMessage)
	}

	reqBody := groqChatRequest{
		Model: model(),
		Messages: []groqChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", errors.New(genericFailureMessage)
	}

	return send(ctx, payload)
}

// visionContentPart / visionMessage / visionRequest mirror the OpenAI-style
// multimodal message shape Groq's vision models expect (a content array
// mixing text and image_url parts), which the plain groqChatMessage
// (string-only Content) can't represent.
type visionContentPart struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL *struct {
		URL string `json:"url"`
	} `json:"image_url,omitempty"`
}

type visionMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type visionRequest struct {
	Model       string          `json:"model"`
	Messages    []visionMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
	MaxTokens   int             `json:"max_tokens"`
}

// ExtractImageText performs OCR by asking a Groq vision model to transcribe
// the given image (as base64 data) verbatim. mimeType should be the image's
// actual content type, e.g. "image/png" or "image/jpeg".
func ExtractImageText(ctx context.Context, base64Data, mimeType string) (string, error) {
	if apiKey() == "" {
		log.Printf("[ai] GROQ_API_KEY is not set in this process's environment")
		return "", errors.New(genericFailureMessage)
	}

	dataURL := "data:" + mimeType + ";base64," + base64Data

	reqBody := visionRequest{
		Model: visionModel,
		Messages: []visionMessage{
			{
				Role: "user",
				Content: []visionContentPart{
					{Type: "text", Text: ocrSystemPrompt},
					{Type: "image_url", ImageURL: &struct {
						URL string `json:"url"`
					}{URL: dataURL}},
				},
			},
		},
		Temperature: temperature,
		MaxTokens:   ocrMaxTokens,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", errors.New(genericFailureMessage)
	}

	return send(ctx, payload)
}
