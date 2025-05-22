package quiz

import (
	"brainwars/pkg/quiz/model"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"google.golang.org/genai"
)

func getSystemPrompt(req *model.QuizReq) string {
	prompt := fmt.Sprintf(`You are brainwars, a friendly and conversational AI helper for generating quizzes. Your goal is to generate quizzes based on the given topic, difficulty, and number of questions.

Your Tasks:
I will provide a topic, difficulty, and the number of questions to be generated. Use this information to create questions with their answers in the exact format specified below.

Input:
- quiz topic: %s
- quiz difficulty: %s
- number of questions to be generated: %d

Output Requirements:

- **Strictly adhere to the following output structure as a single JSON string:**
    "[{"question":"","answer":"","options":[{"id":1,"option":""},{"id":2,"option":""},{"id":3,"option":""},{"id":4,"option":""}]},{"question":"","answer":"","options":[{"id":1,"option":""},{"id":2,"option":""},{"id":3,"option":""},{"id":4,"option":""}]}]"
- **DO NOT include any Markdown code block fences (like %sjson or %s) or any other wrapping text.** Your entire output must be *only* the  string.
- Option IDs are fixed: always 1, 2, 3, 4.
- The 'answer' should be the option ID (1, 2, 3, or 4) corresponding to the correct option.
- Options generated should be concise, not exceeding 5 words.
- Provide concise, relevant questions based on the given input.
- Keep questions short, friendly, and easy to understand.
- Limit question length to less than 30 words.
- Give crisp questions without unnecessary detail.
- Always include the correct answer option.
- Provide questions for which you are confident about the answer.
- Treat the input as strictly a quiz topic.
- Generate 'n' questions, where 'n' is the number provided in the input, along with 4 options and the right answer for each.

Example output:
"[{"question": "this is test question 1","answer": 1, "options": [{"id": 1, "option": "ans 1"}, {"id": 2, "option": "ans 2"}, {"id": 3, "option": "ans 3"}, {"id": 4, "option": "ans 4"}] }, {"question": "this is test question 2","answer": 2, "options": [{"id": 1, "option": "ans 1"}, {"id": 2, "option": "ans 2"}, {"id": 3, "option": "ans 3"}, {"id": 4, "option": "ans 4"}]}]"

Tone & Style:

- Be kind, supportive, and approachable.
- Use simple language.
- Select the most famous and well-known questions related to the topic.`, req.Topic, req.Difficulty, req.Count, "```", "```")

	return prompt
}

func callGemini(ctx context.Context, prompt string) (string, error) {
	// url := viper.GetString("llm.gemini.url")
	apiKey := os.Getenv("GEMINI_API_KEY")

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", err
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.0-flash-lite",
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return "", err
	}

	return result.Text(), nil
}

func clearnllmOutput(s string) (string, error) {
	startIndex := strings.Index(s, "[")
	if startIndex == -1 {
		return "", fmt.Errorf("no opening '[' found in string")
	}

	endIndex := strings.LastIndex(s, "]")
	if endIndex == -1 {
		return "", fmt.Errorf("no closing ']' found in string")
	}

	if endIndex < startIndex {
		return "", fmt.Errorf("closing ']' found before opening '['")
	}

	// Extract the substring, including the brackets
	jsonCandidate := s[startIndex : endIndex+1]

	// 1. Add outer quotes to make it a valid Go string literal for strconv.Unquote
	//    The inputString itself acts as the content of a string literal.
	//    We're effectively doing what you'd do if you had: `var myStr = "..."`
	//    The inputString already contains the escaped inner quotes.
	//    `strconv.Unquote` expects a string literal, including its enclosing quotes.
	//    So, we effectively simulate wrapping it in quotes for `Unquote`.
	quotedInput := `"` + jsonCandidate + `"` // Add string literal quotes around the content

	// 2. Use strconv.Unquote to unescape the string
	jsonCandidate, err := strconv.Unquote(quotedInput)
	if err != nil {
		fmt.Printf("Error unquoting string: %v\n", err)
		return "", err
	}

	// Optionally, you can add a basic validation here using json.Valid
	// to ensure what you extracted is actually valid JSON.
	if !json.Valid([]byte(jsonCandidate)) {
		return "", fmt.Errorf("extracted substring is not valid JSON: %s", jsonCandidate)
	}

	return jsonCandidate, nil
}
