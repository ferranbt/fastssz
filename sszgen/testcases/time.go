package testcases

import "time"

//go:generate go run ../main.go --path time.go

type TimeType struct {
	Timestamp time.Time
	Int       uint64
}

type TimeRawType struct {
	Timestamp uint64
	Int       uint64
}
