package provider

import (
	"context"
	"math/rand/v2"
	"time"
)

type Fake struct{}

func NewFake() *Fake {
	return &Fake{}
}

func (f *Fake) Completion(ctx context.Context, request string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(200 * time.Millisecond):
		val := rand.IntN(10)
		if val < 2 {
			// emulate error
			return "", ErrTooManyRequests
		} else {
			// emulate success
			return "fake resposne", nil
		}
	}
}
