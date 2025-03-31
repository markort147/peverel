package main

type Data interface {
	Init(string)
	AddTask(*Task) TaskId
	AddGroup(*Group) GroupId
	AddRelation(GroupId, ...TaskId) error
	CompleteTask(TaskId) error
	GetTasksByGroup(id GroupId) map[TaskId]*Task
	GetUnassignedTasks() map[TaskId]*Task
	GetGroups() map[GroupId]*Group
	GetTask(id TaskId) *Task
	GetTasks() map[TaskId]*Task
	UnassignTask(TaskId) error
	DeleteTask(id TaskId) error
	DeleteGroup(id GroupId) error
}

type TaskId int
type GroupId int
