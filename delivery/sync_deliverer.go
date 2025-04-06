package delivery

import (
	"sync/atomic"
	"time"

	"github.com/mwazovzky/cloudlog/client"
)

// SyncDeliverer implements LogDeliverer for synchronous delivery
// It's essentially a thin wrapper around a LogSender
type SyncDeliverer struct {
	sender    client.LogSender
	delivered uint64
	failed    uint64
}

// NewSyncDeliverer creates a new synchronous deliverer
func NewSyncDeliverer(sender client.LogSender) *SyncDeliverer {
	return &SyncDeliverer{
		sender: sender,
	}
}

// Deliver sends the log entry synchronously
func (d *SyncDeliverer) Deliver(job string, level string, message string, formatted []byte, timestamp time.Time) error {
	err := d.sender.Send(job, formatted)
	if err != nil {
		atomic.AddUint64(&d.failed, 1)
		return err
	}

	atomic.AddUint64(&d.delivered, 1)
	return nil
}

// Flush is a no-op for synchronous delivery
func (d *SyncDeliverer) Flush() error {
	return nil
}

// Close is a no-op for synchronous delivery
func (d *SyncDeliverer) Close() error {
	return nil
}

// Status returns current delivery statistics
func (d *SyncDeliverer) Status() DeliveryStatus {
	return DeliveryStatus{
		Delivered: int(atomic.LoadUint64(&d.delivered)),
		Failed:    int(atomic.LoadUint64(&d.failed)),
	}
}
