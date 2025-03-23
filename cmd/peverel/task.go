package main

import "time"

type Task struct {
	Name          string
	Description   string
	Period        int
	LastCompleted time.Time
}
