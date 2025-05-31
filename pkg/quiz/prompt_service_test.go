package quiz

import (
	"brainwars/pkg/quiz/model"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func setviper() error {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath("../../config/") // path to look for the config file in

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	err = godotenv.Load("../../" + viper.GetString("app.env"))
	return err
}
func TestOpenrouter(t *testing.T) {
	err := setviper()
	if err != nil {
		t.Error(err.Error())
	}
	req := &model.QuizReq{
		Topic:      "cricket",
		Count:      10,
		Difficulty: "easy",
	}
	systemPrompt := getSystemPrompt(req)
	llmResponse, err := callOpenRouter(context.TODO(), systemPrompt, req)
	if err != nil {
		t.Error(err.Error())
	}
	resp := model.OpenRouterChatResponse{}
	if err := json.Unmarshal([]byte(llmResponse), &resp); err != nil {
		t.Error(err.Error())
	}
	fmt.Println(resp.Choices[0].Message.Content)
	finalresult := resp.Choices[0].Message.Content
	questData := []*model.QuestionData{}
	// sample
	err = json.Unmarshal([]byte(finalresult), &questData)
	if err != nil {
		t.Error(err.Error())
	}
	fmt.Println(finalresult)
}
