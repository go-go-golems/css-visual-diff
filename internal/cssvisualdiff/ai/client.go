package ai

import (
	"context"
	"errors"
)

type Answer struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	Error      string  `json:"error"`
}

type Client interface {
	AnswerQuestion(ctx context.Context, imagePath string, question string) (Answer, error)
}

type NoopClient struct{}

func (n NoopClient) AnswerQuestion(ctx context.Context, imagePath string, question string) (Answer, error) {
	msg := "ai client not configured"
	return Answer{Text: "", Confidence: 0, Error: msg}, errors.New(msg)
}
