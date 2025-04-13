package openai

import "golang.org/x/time/rate"

var limiter = rate.NewLimiter(1, 5) // 1 request per second with burst of 5

type RequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIRequest struct {
	Model       string           `json:"model"`
	Messages    []RequestMessage `json:"messages"`
	Temperature float64          `json:"temperature"`
}

type ResponseMessage struct {
	Content string `json:"content"`
}

type Choice struct {
	Message ResponseMessage `json:"message"`
}

type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
}
