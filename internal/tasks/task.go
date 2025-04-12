package tasks

import "time"

type Task struct {
	Id            TaskId
	Name          string
	Description   string
	Period        int
	LastCompleted time.Time
}

type TaskId int
