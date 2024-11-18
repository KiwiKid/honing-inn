package main

import "time"

func getTodayStr() string {
	return time.Now().Format("Monday, January 2, 2006 at 3:04:05 PM")
}
