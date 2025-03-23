package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/markort147/gopkg/log"
	"net/http"
	"strconv"
	"time"
)

func GetNewTaskForm(c echo.Context) error {
	return c.Render(http.StatusOK, "new-task", nil)
}

func GetNewGroupForm(c echo.Context) error {
	return c.Render(http.StatusOK, "new-group", nil)
}

func PostTask(c echo.Context) error {
	period, _ := strconv.Atoi(c.FormValue("period"))
	data.AddTask(&Task{
		Name:        c.FormValue("name"),
		Description: c.FormValue("description"),
		Period:      period,
	})
	return GetGroups(c)
}

func PostGroup(c echo.Context) error {
	data.AddGroup(&Group{
		Name: c.FormValue("name"),
	})
	return c.NoContent(http.StatusOK)
}

func getTasksByGroup(groupId int) map[TaskId]*Task {
	tasks := make(map[TaskId]*Task)
	relations := data.Relations[GroupId(groupId)]

	if relations != nil && len(relations) > 0 {
		for taskId, task := range data.Tasks {
			for _, relation := range relations {
				if taskId == relation {
					tasks[taskId] = task
				}
			}
		}
	}

	return tasks
}

func getUnassignedTasks() map[TaskId]*Task {

	unassignedTasks := make(map[TaskId]*Task)

	for taskId, task := range data.Tasks {
		isAssigned := false

		for _, relation := range data.Relations {
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

func GetTasks(c echo.Context) error {
	tasks := make(map[TaskId]*Task)
	groupId := c.QueryParam("group")
	switch groupId {
	case "":
		tasks = data.Tasks
	case "none":
		tasks = getUnassignedTasks()
	default:
		intGroupId, _ := strconv.Atoi(groupId)
		tasks = getTasksByGroup(intGroupId)
	}

	template := "tasks-table"
	layout := c.QueryParam("layout")
	if layout == "options" {
		template = "tasks-options"
	}

	return c.Render(http.StatusOK, template, tasks)
}

func GetGroups(c echo.Context) error {
	return c.Render(http.StatusOK, "groups", data.Groups)
}

func PutTaskComplete(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	taskId := TaskId(id)
	_ = data.CompleteTask(taskId)
	return c.HTML(http.StatusOK, renderTaskNextTime(taskId))
}

func CreateMockTasks(c echo.Context) error {
	weekAgo := time.Now().Add(-24 * 7 * time.Hour)
	t0 := data.AddTask(&Task{
		Name:          "vacuum living room",
		Description:   "vacuuming the carpet of the living room",
		Period:        7,
		LastCompleted: weekAgo,
	})

	feb, _ := time.Parse("2006-01-02", "2025-02-01")
	t1 := data.AddTask(&Task{
		Name:          "dust living room",
		Description:   "remove dust from living room furnitures",
		Period:        20,
		LastCompleted: feb,
	})

	jan, _ := time.Parse("2006-01-02", "2025-01-01")
	t2 := data.AddTask(&Task{
		Name:          "wash the wc",
		Description:   "wash the wc with bleach",
		Period:        1,
		LastCompleted: jan,
	})

	living := data.AddGroup(&Group{
		Name: "Living Room",
	})
	bathroom := data.AddGroup(&Group{
		Name: "Bath Room",
	})

	data.AddRelation(living, t0, t1)
	data.AddRelation(bathroom, t2)

	data.AddTask(&Task{
		Name:          "clean the bike",
		Description:   "clean the chain of the bike",
		Period:        30,
		LastCompleted: jan,
	})

	return GetGroups(c)
}

func GetTaskNextTime(c echo.Context) error {
	taskId, _ := strconv.Atoi(c.Param("id"))
	return c.HTML(http.StatusOK, renderTaskNextTime(TaskId(taskId)))
}

func renderTaskNextTime(taskId TaskId) string {
	task := data.Tasks[taskId]

	nextDay := task.LastCompleted.AddDate(0, 0, task.Period)

	layout := "20060102"
	todayStr := time.Now().Format(layout)
	nextDayStr := nextDay.Format(layout)

	if todayStr == nextDayStr {
		return "today"
	}
	todayTime, _ := time.Parse(layout, todayStr)
	nextDayTime, _ := time.Parse(layout, nextDayStr)
	diff := int(nextDayTime.Sub(todayTime).Hours() / 24)
	if diff < 0 {
		return fmt.Sprintf("%d days ago", -diff)
	}
	return fmt.Sprintf("%d days", diff)
}

func PutGroupAssignTask(c echo.Context) error {
	groupId, _ := strconv.Atoi(c.Param("id"))
	taskId, _ := strconv.Atoi(c.FormValue("assign-task"))

	err := data.AddRelation(GroupId(groupId), TaskId(taskId))
	if err != nil {
		log.Logger.Errorf("add group assign task err: %v", err)
	}

	return c.Render(http.StatusOK, "groups", data.Groups)
}
