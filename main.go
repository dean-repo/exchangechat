package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"gopkg.in/fsnotify.v1"
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

func index(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title": "Please provide the OpenAI API KEY",
	})
}

func main() {
	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

	router.GET("/", index)
	router.POST("/sendMessage", sendMessage)
	router.GET("/json", welcome)

	router.Static("/public", "./public")

	// creates a new file watcher for App_offline.htm
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("ERROR", err)
	}
	defer watcher.Close()

	// watch for App_offline.htm and exit the program if present
	// This allows continuous deployment on App Service as the .exe will not be
	// terminated otherwise
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if strings.HasSuffix(event.Name, "app_offline.htm") {
					fmt.Println("Exiting due to app_offline.htm being present")
					os.Exit(0)
				}
			}
		}
	}()

	// get the current working directory and watch it
	currentDir, err := os.Getwd()
	if err := watcher.Add(currentDir); err != nil {
		fmt.Println("ERROR", err)
	}

	port := os.Getenv("HTTP_PLATFORM_PORT")
	if port == "" {
		port = "8080"
	}

	router.Run("127.0.0.1:" + port)
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
