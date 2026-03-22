package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/mwazovzky/cloudlog/client"
)

// SyncSender delivers log entries synchronously via a LogSender
type SyncSender struct {
	client client.LogSender
}

// NewSyncSender creates a sender that delivers immediately via the given LogSender
func NewSyncSender(client client.LogSender) *SyncSender {
	return &SyncSender{client: client}
}

// Send builds a LokiEntry from the formatted content and sends it immediately
func (s *SyncSender) Send(ctx context.Context, content []byte, labels map[string]string, timestamp time.Time) error {
	entry := client.LokiEntry{
		Streams: []client.LokiStream{
			{
				Stream: labels,
				Values: [][]string{
					{
						fmt.Sprintf("%d", timestamp.UnixNano()),
						string(content),
					},
				},
			},
		},
	}

	return s.client.Send(ctx, entry)
}
