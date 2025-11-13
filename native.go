package main

import "time"

var startTime = time.Now()

func ClockNative(argCount int, args *Value) Value {
	elapsed := time.Since(startTime)
	seconds := elapsed.Seconds()
	return NewFloat(seconds)
}
