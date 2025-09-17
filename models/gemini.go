package models

import (
	"context"
	"fmt"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

type Gemini struct {
	Message string
	System  string
}

func NewGemini(system, message string) *Gemini {
	return &Gemini{
		Message: message,
		System:  system,
	}
}

func (g *Gemini) SendMessage() (string, error) {
	env, err := godotenv.Read(".env")
	if err != nil {
		return "", fmt.Errorf("error reading .env file from current directory: %w", err)
	}

	apiKey, ok := env["GEMINI_API_KEY"]
	if !ok || apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY not found or is empty in .env file")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", err
	}

	resp, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", genai.Text(g.System+"\n"+g.Message), nil)
	if err != nil {
		return "", err
	}

	if resp.Candidates[0].Content != nil {
		return resp.Text(), nil
	}

	return "", fmt.Errorf("no content found in response")
}
