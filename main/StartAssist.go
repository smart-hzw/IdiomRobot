package main

import (
	"errors"
	"math"
	"time"
)

const concurrencyTimeWindowSec = 2

func CheckSessionLimit(wssInfo *WssInfo) error {
	if wssInfo.Shards > wssInfo.SessionStartLimit.Remaining {
		return errors.New("Session limit reached")
	}
	return nil
}

// 计算并发时间间隔
func CalcInterval(maxConcurrency uint32) time.Duration {
	if maxConcurrency == 0 {
		maxConcurrency = 1
	}
	f := math.Round(concurrencyTimeWindowSec / float64(maxConcurrency))
	if f == 0 {
		f = 1
	}
	return time.Duration(f) * time.Second
}
