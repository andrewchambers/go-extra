package conc

import (
	"context"
	"testing"
)

func TestBundle(t *testing.T) {
	b := NewBundle(context.Background())

	c1 := make(chan struct{}, 1)
	c2 := make(chan struct{}, 1)

	b.Go(func(ctx context.Context) {
		select {
		case <-ctx.Done():
			c1 <- struct{}{}
			return
		}

	})

	b.Go(func(ctx context.Context) {
		select {
		case <-ctx.Done():
			c2 <- struct{}{}
			return
		}
	})

	b.Close()

	select {
	case <-c1:
	default:
		t.FailNow()
	}

	select {
	case <-c2:
	default:
		t.FailNow()
	}
}
