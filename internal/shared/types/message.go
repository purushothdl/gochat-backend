package types

import "time"

// ReceiptInfo contains the user and the timestamp for a specific receipt.
type ReceiptInfo struct {
	User      *BasicUser `json:"user"`
	Timestamp time.Time  `json:"timestamp"`
}