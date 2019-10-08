package main

import (
	"context"
	"testing"
)

func Test(t *testing.T) {
	m := PubSubMessage{
		Data: []byte(""),
	}
	SyncStarling(context.Background(), m)
}
