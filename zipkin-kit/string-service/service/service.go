package service

import (
	"context"
	"errors"
	"strings"
)

const (
	StrMaxSize = 1024
)

var (
	ErrMaxSize = errors.New("maximum size of 1024 bytes exceeded")

	ErrStrValue = errors.New("error str value to Integer")
)

type Service interface {
	Concat(a, b string) (string, error)
	Diff(ctx context.Context, a, b string) (string, error)
	HealthCheck() bool
}

type StringService struct {
}

func (s StringService) Concat(a, b string) (string, error) {
	if len(a)+len(b) > StrMaxSize {
		return "", ErrMaxSize
	}
	return a + b, nil
}

func (s StringService) Diff(ctx context.Context, a, b string) (string, error) {
	if len(a) < 1 || len(b) < 1 {
		return "", nil
	}
	res := ""
	if len(a) >= len(b) {
		for _, char := range b {
			if strings.Contains(a, string(char)) {
				res = res + string(char)
			}
		}
	} else {
		for _, char := range a {
			if strings.Contains(b, string(char)) {
				res = res + string(char)
			}
		}
	}
	return res, nil
}

func (s StringService) HealthCheck() bool {
	return true
}

//装饰者
type ServiceMiddleware func(Service) Service
