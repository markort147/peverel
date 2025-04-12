package tasks

type Data interface {
	Init(string)
	AddTask(*Task) TaskId
	UpdateTask(TaskId, *Task) error
	AddGroup(*Group) GroupId
	SetRelation(GroupId, ...TaskId) error
	CompleteTask(TaskId) error
	GetTasksByGroup(id GroupId) map[TaskId]*Task
	GetUnassignedTasks() map[TaskId]*Task
	GetGroups() map[GroupId]*Group
	GetGroup(id GroupId) *Group
	GetTask(id TaskId) *Task
	GetTasks() map[TaskId]*Task
	UnassignTask(TaskId) error
	DeleteTask(id TaskId) error
	DeleteGroup(id GroupId) error
}
