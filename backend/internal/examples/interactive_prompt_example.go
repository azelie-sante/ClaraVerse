package examples

import (
	"github.com/google/uuid"

	"claraverse/internal/models"
)

// ExampleSimplePrompt shows how to create a simple interactive prompt
// with basic question types (text, number)
func ExampleSimplePrompt(conversationID string) models.ServerMessage {
	return models.ServerMessage{
		Type:           "interactive_prompt",
		PromptID:       uuid.New().String(),
		ConversationID: conversationID,
		Title:          "Need More Information",
		Description:    "To help you better, I need a few more details.",
		Questions: []models.InteractiveQuestion{
			{
				ID:          "name",
				Type:        "text",
				Label:       "What's your name?",
				Placeholder: "Enter your name...",
				Required:    true,
			},
			{
				ID:       "age",
				Type:     "number",
				Label:    "How old are you?",
				Required: false,
				Validation: &models.QuestionValidation{
					Min: floatPtr(0),
					Max: floatPtr(150),
				},
			},
		},
		AllowSkip: boolPtr(true),
	}
}

// ExampleComplexPrompt shows how to create a complex prompt
// with all question types and validation
func ExampleComplexPrompt(conversationID string) models.ServerMessage {
	return models.ServerMessage{
		Type:           "interactive_prompt",
		PromptID:       uuid.New().String(),
		ConversationID: conversationID,
		Title:          "Create a New Project",
		Description:    "To create your project, I need some information about your requirements.",
		Questions: []models.InteractiveQuestion{
			{
				ID:         "language",
				Type:       "select",
				Label:      "What programming language do you want to use?",
				Required:   true,
				Options:    []string{"Python", "JavaScript", "TypeScript", "Java", "Go"},
				AllowOther: true,
			},
			{
				ID:         "features",
				Type:       "multi-select",
				Label:      "Which features do you need?",
				Required:   true,
				Options:    []string{"Authentication", "Database", "API", "Testing"},
				AllowOther: true,
			},
			{
				ID:       "complexity",
				Type:     "number",
				Label:    "Complexity level (1-10)",
				Required: true,
				Validation: &models.QuestionValidation{
					Min: floatPtr(1),
					Max: floatPtr(10),
				},
			},
			{
				ID:       "async",
				Type:     "checkbox",
				Label:    "Use async/await?",
				Required: false,
			},
			{
				ID:          "description",
				Type:        "text",
				Label:       "Project description",
				Placeholder: "Describe your project...",
				Required:    false,
				Validation: &models.QuestionValidation{
					MinLength: intPtr(10),
					MaxLength: intPtr(200),
				},
			},
		},
		AllowSkip: boolPtr(false), // User must answer
	}
}

// ExampleEmailValidation shows how to create a prompt with email validation
func ExampleEmailValidation(conversationID string) models.ServerMessage {
	return models.ServerMessage{
		Type:           "interactive_prompt",
		PromptID:       uuid.New().String(),
		ConversationID: conversationID,
		Title:          "Email Verification",
		Description:    "Please verify your email address to continue.",
		Questions: []models.InteractiveQuestion{
			{
				ID:          "email",
				Type:        "text",
				Label:       "Email address",
				Placeholder: "your@email.com",
				Required:    true,
				Validation: &models.QuestionValidation{
					Pattern: `^[^\s@]+@[^\s@]+\.[^\s@]+$`, // Email regex
				},
			},
			{
				ID:       "agree",
				Type:     "checkbox",
				Label:    "I agree to the terms and conditions",
				Required: true,
			},
		},
		AllowSkip: boolPtr(false),
	}
}

// Example of how to use SendInteractivePrompt in a tool or handler
//
// func (h *WebSocketHandler) SomeToolOrHandler(userConn *models.UserConnection) {
//     // Create a prompt
//     prompt := ExampleSimplePrompt(userConn.ConversationID)
//
//     // Send it to the user
//     success := h.SendInteractivePrompt(userConn, prompt)
//     if !success {
//         log.Printf("Failed to send prompt to user")
//         return
//     }
//
//     // The response will be received in handleInteractivePromptResponse
//     // You can store the promptID and wait for the response before continuing
// }

// Example of sending a validation error
func ExampleValidationError(conversationID, promptID string) models.ServerMessage {
	return models.ServerMessage{
		Type:           "prompt_validation_error",
		PromptID:       promptID,
		ConversationID: conversationID,
		Errors: map[string]string{
			"email": "Please enter a valid email address",
			"age":   "Age must be between 0 and 150",
		},
	}
}

// Example of sending a timeout message
func ExampleTimeout(conversationID, promptID string) models.ServerMessage {
	return models.ServerMessage{
		Type:           "prompt_timeout",
		PromptID:       promptID,
		ConversationID: conversationID,
		ErrorMessage:   "Prompt timed out. Please try again.",
	}
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}
