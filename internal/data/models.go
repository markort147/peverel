package data

import "time"

type Task struct {
	Id            TaskId
	Name          string
	Description   string
	Period        int
	LastCompleted time.Time
}

type TaskId int

/*type Group struct {
	Id   GroupId
	Name string
}*/

// type GroupId int
