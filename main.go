package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func sendMessage(c *gin.Context) {
	var incomingData PromptBody
	client := http.Client{}

	if err := c.BindJSON(&incomingData); err != nil {
		log.Println("Error: " + err.Error())
	}

	if len(incomingData.Model) == 0 {
		incomingData.Model = "text-davinci-003"
	}

	if len(incomingData.MaxTokens) == 0 {
		incomingData.MaxTokens = "16"
	}

	if len(incomingData.Temperature) == 0 {
		incomingData.Temperature = "0.5"
	}

	if len(incomingData.Prompt) == 0 {
		return
	}

	payload := strings.NewReader(`{
		"model":"` + incomingData.Model + `",
		"prompt":"` + incomingData.Prompt + `",
		 "max_tokens":` + incomingData.MaxTokens + `,
		 "temperature":` + incomingData.Temperature + `
		}`)

	url := "https://api.openai.com/v1/completions"
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("Authorization", "Bearer "+incomingData.Secret)
	req.Header.Add("Content-Type", "application/json")

	response, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	responseData, _ := io.ReadAll(response.Body)

	var results OpenApiRes

	json.Unmarshal([]byte(responseData), &results)

	c.IndentedJSON(http.StatusOK, results)

}

func welcome(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"foo": "bar",
	})
}

func main() {
	router := gin.Default()
	router.POST("/sendMessage", sendMessage)
	router.GET("/", welcome)
	router.Run("127.0.0.1:8080")
}

type PromptBody struct {
	Model       string `json:"model"`
	Prompt      string `json:"prompt"`
	MaxTokens   string `json:"max_tokens"`
	Temperature string `json:"temperature"`
	Secret      string `json:"secret"`
}

type OpenApiRes struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"Created"`
	Model   string `json:"Model"`
	Choices []struct {
		Text         string `json:"text"`
		Index        int    `json:"index"`
		Logprobs     string `json:"Logprobs"`
		FinishReason string `json:"finish_reason"`
	} `json:"Choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"Usage"`
}
