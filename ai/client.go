package ai

import (
	"context"
	"encoding/json"

	"github.com/openai/openai-go"
)

/*
Conn represents a connection to the OpenAI API.
It manages the client and conversation history for generating optimization suggestions.
*/
type Conn struct {
	client  *openai.Client
	history []openai.ChatCompletionMessageParamUnion
}

/*
NewConn creates a new OpenAI API connection.
It initializes the OpenAI client for making API requests.
*/
func NewConn() *Conn {
	return &Conn{client: openai.NewClient()}
}

/*
Generate sends a prompt to the OpenAI API and returns the generated optimization suggestion.
It configures the API request with the appropriate schema and model settings.
*/
func (conn *Conn) Generate(
	ctx context.Context,
	prompt *Prompt,
) (any, error) {
	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.F("optimization_suggestion"),
		Description: openai.F("A detailed optimization suggestion for a MongoDB database"),
		Schema:      openai.F(prompt.schema),
		Strict:      openai.Bool(true),
	}

	chat, err := conn.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt.user),
		}),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type:       openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(schemaParam),
			},
		),
		Model: openai.F(openai.ChatModelGPT4o),
	})

	if err != nil {
		return "", err
	}

	var result any

	if err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &result); err != nil {
		return "", err
	}

	return result, nil
}
