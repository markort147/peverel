package main

import (
	"fmt"
	"time"
)

type MemoryData struct {
	Tasks       map[TaskId]*Task
	Groups      map[GroupId]*Group
	Relations   map[GroupId][]TaskId
	nextTaskId  TaskId
	nextGroupId GroupId
}

func (m *MemoryData) GetGroups() map[GroupId]*Group {
	return m.Groups
}

func (m *MemoryData) Init(_ string) {
	m.Tasks = make(map[TaskId]*Task)
	m.Groups = make(map[GroupId]*Group)
	m.Relations = make(map[GroupId][]TaskId)
}

func (m *MemoryData) AddTask(task *Task) TaskId {
	id := m.nextTaskId
	m.Tasks[id] = task
	m.nextTaskId++
	Logger.Infof("add task [%d] %+v", id, task)
	return id
}

func (m *MemoryData) AddGroup(group *Group) GroupId {
	id := m.nextGroupId
	m.Groups[id] = group
	m.nextGroupId++
	Logger.Infof("add group [%d] %+v", id, group)
	return id
}

func (m *MemoryData) AddRelation(groupId GroupId, taskIds ...TaskId) error {
	Logger.Debugf("adding tasks %+v to group %d", groupId, taskIds)

	_, gExists := m.Groups[groupId]

	if !gExists {
		return fmt.Errorf("group %d does not exist", groupId)
	}

	for _, taskId := range taskIds {
		if _, tExists := m.Tasks[taskId]; !tExists {
			return fmt.Errorf("task %d does not exist", taskId)
		}
	}

	_, rExists := m.Relations[groupId]
	if !rExists {
		m.Relations[groupId] = make([]TaskId, 0)
	}
	for _, taskId := range taskIds {
		m.Relations[groupId] = append(m.Relations[groupId], taskId)
	}

	Logger.Debugf("updated relation: [%d] %+v", groupId, m.Relations[groupId])

	return nil
}

func (m *MemoryData) CompleteTask(taskId TaskId) error {
	if task, exists := m.Tasks[taskId]; exists {
		task.LastCompleted = time.Now()
		return nil
	}
	return fmt.Errorf("task %d does not exist", taskId)
}

func (m *MemoryData) GetTasksByGroup(groupId GroupId) map[TaskId]*Task {
	tasks := make(map[TaskId]*Task)
	relations := m.Relations[groupId]

	if relations != nil && len(relations) > 0 {
		for taskId, task := range m.Tasks {
			for _, relation := range relations {
				if taskId == relation {
					tasks[taskId] = task
				}
			}
		}
	}

	return tasks
}

func (m *MemoryData) GetUnassignedTasks() map[TaskId]*Task {

	unassignedTasks := make(map[TaskId]*Task)

	for taskId, task := range m.Tasks {
		isAssigned := false

		for _, relation := range m.Relations {
			for _, relatedTaskId := range relation {
				if taskId == relatedTaskId {
					isAssigned = true
					break
				}
			}
			if isAssigned {
				break
			}
		}

		if !isAssigned {
			unassignedTasks[taskId] = task
		}
	}

	return unassignedTasks
}

func (m *MemoryData) GetTasks() map[TaskId]*Task {
	return m.Tasks
}

func (m *MemoryData) GetTask(id TaskId) *Task {
	return m.Tasks[id]
}
