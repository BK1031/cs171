package llm

import (
	"context"
	"fmt"
	"server/config"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

var Client *genai.Client
var Model *genai.GenerativeModel

var prefix = "You are given chat history in the form of Query: <query> and Answer: <answer>. Please answer the latest query and return a single line answer with no prefix.\n\n"

func Initialize(modelName string) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(config.GeminiAPIKey))
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}

	Client = client
	Model = client.GenerativeModel(modelName)
}

func Query(query string) (string, error) {
	// Generate content
	ctx := context.Background()
	resp, err := Model.GenerateContent(ctx, genai.Text(prefix+query))
	if err != nil {
		fmt.Printf("Error generating content: %v\n", err)
		return "", err
	}

	return string(resp.Candidates[0].Content.Parts[0].(genai.Text)), nil
}
