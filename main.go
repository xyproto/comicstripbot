package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/robfig/cron/v3"
)

const (
	dalleAPIURL  = "https://api.openai.com/v1/images/generations"
	slackAPIURL  = "https://slack.com/api/chat.postMessage"
	slackChannel = "#ai-comic-strips" // Updated Slack channel
)

type DalleRequest struct {
	Prompt string `json:"prompt"`
}

type DalleResponse struct {
	Data []struct {
		Urls []string `json:"urls"`
	} `json:"data"`
}

type SlackMessage struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

func generateImage() (string, error) {
	client := &http.Client{}
	prompt := "A knock-knock cartoon strip" // Modify this prompt as needed

	reqBody, _ := json.Marshal(DalleRequest{Prompt: prompt})
	req, _ := http.NewRequest("POST", dalleAPIURL, bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var dalleResp DalleResponse
	if err := json.NewDecoder(resp.Body).Decode(&dalleResp); err != nil {
		return "", err
	}

	if len(dalleResp.Data) > 0 && len(dalleResp.Data[0].Urls) > 0 {
		return dalleResp.Data[0].Urls[0], nil
	}

	return "", fmt.Errorf("no image generated")
}

func postToSlack(imageURL string) error {
	msg := SlackMessage{
		Channel: slackChannel,
		Text:    fmt.Sprintf("Here's your knock-knock cartoon for the day! %s", imageURL),
	}

	jsonData, _ := json.Marshal(msg)
	req, _ := http.NewRequest("POST", slackAPIURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SLACK_BOT_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("Slack response: %s\n", string(body))
	return nil
}

func main() {
	c := cron.New()
	c.AddFunc("0 9,12,15 * * *", func() {
		imageURL, err := generateImage()
		if err != nil {
			fmt.Printf("Error generating image: %s\n", err)
			return
		}
		if err := postToSlack(imageURL); err != nil {
			fmt.Printf("Error posting to Slack: %s\n", err)
		}
	})
	c.Start()

	// Keep the application running
	select {}
}
