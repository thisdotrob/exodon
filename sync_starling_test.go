package starling

import (
	"context"
	"testing"
)

func TestSyncStarling(t *testing.T) {
	m := PubSubMessage{
		Data: []byte(""),
	}
	SyncStarling(context.Background(), m)
}
