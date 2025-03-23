package main

import (
	"fmt"
	"github.com/markort147/gopkg/log"
	"time"
)

type TaskId int
type GroupId int
type Data struct {
	Tasks       map[TaskId]*Task
	Groups      map[GroupId]*Group
	Relations   map[GroupId][]TaskId
	nextTaskId  TaskId
	nextGroupId GroupId
}

func NewData() Data {
	return Data{
		Tasks:     make(map[TaskId]*Task),
		Groups:    make(map[GroupId]*Group),
		Relations: make(map[GroupId][]TaskId),
	}
}

func (d *Data) AddTask(task *Task) TaskId {
	id := d.nextTaskId
	d.Tasks[id] = task
	d.nextTaskId++
	log.Logger.Infof("add task [%d] %+v", id, task)
	return id
}

func (d *Data) AddGroup(group *Group) GroupId {
	id := d.nextGroupId
	d.Groups[id] = group
	d.nextGroupId++
	log.Logger.Infof("add group [%d] %+v", id, group)
	return id
}

func (d *Data) AddRelation(groupId GroupId, taskIds ...TaskId) error {
	log.Logger.Debugf("adding tasks %+v to group %d", groupId, taskIds)

	_, gExists := d.Groups[groupId]

	if !gExists {
		return fmt.Errorf("group %d does not exist", groupId)
	}

	for _, taskId := range taskIds {
		if _, tExists := d.Tasks[taskId]; !tExists {
			return fmt.Errorf("task %d does not exist", taskId)
		}
	}

	_, rExists := d.Relations[groupId]
	if !rExists {
		d.Relations[groupId] = make([]TaskId, 0)
	}
	for _, taskId := range taskIds {
		d.Relations[groupId] = append(d.Relations[groupId], taskId)
	}

	log.Logger.Debugf("updated relation: [%d] %+v", groupId, d.Relations[groupId])

	return nil
}

func (d *Data) CompleteTask(taskId TaskId) error {
	if task, exists := d.Tasks[taskId]; exists {
		task.LastCompleted = time.Now()
		return nil
	}
	return fmt.Errorf("task %d does not exist", taskId)
}
