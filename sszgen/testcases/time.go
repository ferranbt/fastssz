package testcases

import "time"

type TimeType struct {
	Timestamp time.Time
	Int       uint64
}

type TimeRawType struct {
	Timestamp uint64
	Int       uint64
}
