package dto

import "time"

type RateLimitInfo struct {
	Allowed      bool       `json:"allowed"`
	Remaining    int        `json:"remaining"`
	ResetTime    *time.Time `json:"reset_time,omitempty"`
	BlockedUntil *time.Time `json:"blocked_until,omitempty"`
}
