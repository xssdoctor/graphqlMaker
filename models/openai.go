package models

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type OpenAI struct {
	Message string
	ApiKey string
	System string
}

func NewOpenAi(apikey string, system string, message string) *OpenAI {
	return &OpenAI{
		Message: message,
		System: system,
		ApiKey: apikey,
	}
}	

func (oai *OpenAI) SendMessage() (string, error) {
	messages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleSystem,
			Content: oai.System,
		},
		{
			Role: openai.ChatMessageRoleUser,
			Content: oai.Message,
		},
	}
	client := openai.NewClient(oai.ApiKey)

	ctx := context.Background()
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: "gpt-4o-2024-05-13",
			Messages: messages,
		},
	)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}