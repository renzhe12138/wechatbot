package gtp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/869413421/wechatbot/config"
	"io/ioutil"
	"log"
	"net/http"
)

const BASEURL = "https://api.openai.com/v1/"

// ChatGPTResponseBody 请求体
type ChatGPTResponseBody struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int                    `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChoiceItem           `json:"choices"`
	Usage   map[string]interface{} `json:"usage"`
}
type ChatGPTCheckResponseBody struct {
	ID string `json:"id"`
	Model string `json:"model"`
	Results []Results `json:"results"`
}
type Categories struct {
	Hate bool `json:"hate"`
	HateThreatening bool `json:"hate/threatening"`
	SelfHarm bool `json:"self-harm"`
	Sexual bool `json:"sexual"`
	SexualMinors bool `json:"sexual/minors"`
	Violence bool `json:"violence"`
	ViolenceGraphic bool `json:"violence/graphic"`
}
type CategoryScores struct {
	Hate float64 `json:"hate"`
	HateThreatening float64 `json:"hate/threatening"`
	SelfHarm float64 `json:"self-harm"`
	Sexual float64 `json:"sexual"`
	SexualMinors float64 `json:"sexual/minors"`
	Violence float64 `json:"violence"`
	ViolenceGraphic float64 `json:"violence/graphic"`
}
type Results struct {
	Categories Categories `json:"categories"`
	CategoryScores CategoryScores `json:"category_scores"`
	Flagged bool `json:"flagged"`
}

type ChoiceItem struct {
	Text         string `json:"text"`
	Index        int    `json:"index"`
	Logprobs     int    `json:"logprobs"`
	FinishReason string `json:"finish_reason"`
}

// ChatGPTRequestBody 响应体
type ChatGPTRequestBody struct {
	Model            string  `json:"model"`
	Prompt           string  `json:"prompt"`
	MaxTokens        int     `json:"max_tokens"`
	Temperature      float32 `json:"temperature"`
	TopP             int     `json:"top_p"`
	FrequencyPenalty int     `json:"frequency_penalty"`
	PresencePenalty  int     `json:"presence_penalty"`
}

type ChatGPTCheckRequestBody struct {
	Input            string  `json:"input"`
}

// Completions gtp文本模型回复
//curl https://api.openai.com/v1/completions
//-H "Content-Type: application/json"
//-H "Authorization: Bearer your chatGPT key"
//-d '{"model": "text-davinci-003", "prompt": "give me good song", "temperature": 0, "max_tokens": 7}'
func Completions(msg string) (string, error) {
	requestBody := ChatGPTRequestBody{
		Model:            "text-davinci-003",
		Prompt:           msg,
		MaxTokens:        2048,
		Temperature:      0.9,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
	}
	requestData, err := json.Marshal(requestBody)

	if err != nil {
		return "", err
	}
	log.Printf("request gtp json string : %v", string(requestData))
	req, err := http.NewRequest("POST", BASEURL+"completions", bytes.NewBuffer(requestData))
	if err != nil {
		return "", err
	}

	apiKey := config.LoadConfig().ApiKey
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("gtp api status code not equals 200,code is %d", response.StatusCode))
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	gptResponseBody := &ChatGPTResponseBody{}
	log.Println(string(body))
	err = json.Unmarshal(body, gptResponseBody)
	if err != nil {
		return "", err
	}

	var reply string
	if len(gptResponseBody.Choices) > 0 {
		reply = gptResponseBody.Choices[0].Text
	}
	log.Printf("gpt response text: %s \n", reply)
	return reply, nil
}
func check(msg string) (bool, error) {
	requestBody := ChatGPTCheckRequestBody{
		Input:           msg,
	}
	requestData, err := json.Marshal(requestBody)
	if err != nil {
		return false, err
	}
	req, err := http.NewRequest("POST", BASEURL+"moderations", bytes.NewBuffer(requestData))
	if err != nil {
		return false, err
	}

	apiKey := config.LoadConfig().ApiKey
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return false, errors.New(fmt.Sprintf("gtp api status code not equals 200,code is %d", response.StatusCode))
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, err
	}

	gptResponseBody := &ChatGPTCheckResponseBody{}
	log.Println(string(body))
	err = json.Unmarshal(body, gptResponseBody)
	if err != nil {
		return false, err
	}

	var reply bool
	if len(gptResponseBody.Results) > 0 {
		reply = gptResponseBody.Results[0].Flagged
	}
	log.Printf("gpt response check: %s \n", reply)
	return reply, nil
}
