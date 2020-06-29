package svc2

import (
	"context"
	"errors"
)

const (
	Int64Max = 1<<63 - 1
	Int64Min = -(Int64Max + 1)
)

var ErrIntOverFlow = errors.New("integer overflow occurred")

type Service interface {
	Sum(ctx context.Context, a, b int64) (int64, error)
}
