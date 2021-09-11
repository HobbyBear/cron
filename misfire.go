package cron

import "time"

type FireStorage interface {
	GetEntry(taskId string) *Entry
	PutEntry(taskId string, nextTime time.Time)
	PutRetryEntry(taskId string)
	GetRetryEntryList() []string
	DelRetryEntry(taskId string)
}
