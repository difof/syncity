package syncity

import (
	"context"
	"testing"
	"time"
)

func TestJobQueueWait(t *testing.T) {
	jq := NewJobQueue(context.Background(), 10)

	jq.Enqueue(func(ctx context.Context, arg any) error {
		time.Sleep(2 * time.Second)
		t.Log("Enqueue done")
		return nil
	})

	jq.EnqueueWait(func(ctx context.Context, arg any) error {
		t.Log("EnqueueWait done")
		return nil
	})
}
