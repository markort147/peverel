package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	data "github.com/markor147/peverel/internal/data"
	"github.com/markor147/peverel/internal/log"
)

func GetLayoutHome(c echo.Context) error {
	return c.Render(http.StatusOK, "layout", map[string]string{
		"Title":   "peverel - home",
		"Content": "home",
	})
}

func GetPageHome(c echo.Context) error {
	return c.Render(http.StatusOK, "page/home", nil)
}

func GetLayoutSettings(c echo.Context) error {
	return c.Render(http.StatusOK, "layout", map[string]string{
		"Title":   "peverel - settings",
		"Content": "settings",
	})
}

func GetPageSettings(c echo.Context) error {
	return c.Render(http.StatusOK, "page/settings", nil)
}

func GetLayoutAddTask(c echo.Context) error {
	return c.Render(http.StatusOK, "layout", map[string]string{
		"Title":   "peverel - new task",
		"Content": "add-task",
	})
}

func GetPageAddTask(c echo.Context) error {
	return c.Render(http.StatusOK, "page/add-task", nil)
}

func GetLayoutEditTask(c echo.Context) error {
	taskId := c.QueryParam("id")
	return c.Render(http.StatusOK, "layout", map[string]string{
		"Title":   "peverel - edit task " + taskId,
		"Content": "edit-task?id=" + taskId,
	})
}

func GetPageEditTask(c echo.Context) error {
	taskId, _ := strconv.Atoi(c.QueryParam("id"))
	task, _ := data.GetTask(data.TaskId(taskId))
	return c.Render(http.StatusOK, "page/edit-task", task)
}

func GetNewTaskForm(c echo.Context) error {
	return c.Render(http.StatusOK, "task-form", nil)
}

func GetNewGroupForm(c echo.Context) error {
	return c.Render(http.StatusOK, "new-group", nil)
}

func PostTask(c echo.Context) error {
	period, _ := strconv.Atoi(c.FormValue("period"))
	taskId, err := data.AddTask(data.Task{
		Name:        c.FormValue("name"),
		Description: c.FormValue("description"),
		Period:      period,
	})
	if err != nil {
		return err
	}

	groupId, _ := strconv.Atoi(c.FormValue("group"))
	if groupId != -1 {
		err := data.SetRelation(data.GroupId(groupId), taskId)
		if err != nil {
			return err
		}
	}

	return c.String(http.StatusOK, "task created successfully")
}

func PostGroup(c echo.Context) error {
	if _, err := data.AddGroup(&data.Group{
		Name: c.FormValue("name"),
	}); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

func GetTasks(c echo.Context) error {
	groupId := c.QueryParam("group")
	days := c.QueryParam("days")
	expired := c.QueryParam("expired") != "false"

	log.Logger.Debugf("functions GetTask params: (group: %v, days: %v, expired: %v)", groupId, days, expired)

	tasks, err := data.Tasks(groupId, days, expired)
	if err != nil {
		return err
	}

	layout := c.QueryParam("layout")
	var template string
	switch layout {
	case "options":
		template = "tasks-options"
	case "table":
		template = "tasks-table"
	default:
		template = "tasks-table"
	}

	return c.Render(http.StatusOK, template, tasks)
}

func GetGroups(c echo.Context) error {
	template := "groups"
	layout := c.QueryParam("layout")
	if layout == "options" {
		template = "groups-options"
	}

	groups, err := data.GetGroups()
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, template, groups)
}

func PutTaskComplete(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	taskId := data.TaskId(id)
	_ = data.CompleteTask(taskId)

	tasks, err := data.Tasks("", "", true)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "tasks-table", tasks)
}

func GetTaskNextTime(c echo.Context) error {
	taskId, _ := strconv.Atoi(c.Param("id"))
	return c.HTML(http.StatusOK, renderTaskNextTime(data.TaskId(taskId)))
}

func renderTaskNextTime(taskId data.TaskId) string {
	task, err := data.GetTask(taskId)
	if err != nil {
		log.Logger.Fatalf("renderTaskNextTime: %v", err)
	}

	layout := "20060102"
	nextDay := task.LastCompleted.AddDate(0, 0, task.Period)

	nextDayStr := nextDay.Format(layout)
	nextDayTime, _ := time.Parse(layout, nextDayStr)
	todayStr := time.Now().Format(layout)
	todayTime, _ := time.Parse(layout, todayStr)
	if todayStr == nextDayStr {
		return "today"
	}
	diff := int(nextDayTime.Sub(todayTime).Hours() / 24)
	if diff < 0 {
		return fmt.Sprintf("%d days ago", -diff)
	}
	return fmt.Sprintf("%d days", diff)
}

func PutGroupAssignTask(c echo.Context) error {
	groupId, _ := strconv.Atoi(c.Param("id"))
	taskId, _ := strconv.Atoi(c.FormValue("assign-task"))

	err := data.SetRelation(data.GroupId(groupId), data.TaskId(taskId))
	if err != nil {
		return err
	}

	groups, err := data.GetGroups()
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "groups", groups)
}

func DeleteTask(c echo.Context) error {
	taskId, _ := strconv.Atoi(c.Param("id"))
	if err := data.DeleteTask(data.TaskId(taskId)); err != nil {
		log.Logger.Errorf("Error deleting task: %v", err)
		return err
	}
	return c.String(http.StatusOK, "task deleted successfully")
}

func DeleteGroup(c echo.Context) error {
	groupId, _ := strconv.Atoi(c.Param("id"))
	if err := data.DeleteGroup(data.GroupId(groupId)); err != nil {
		log.Logger.Errorf("Error deleting group: %v", err)
		return err
	}

	groups, err := data.GetGroups()
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "groups", groups)
}

func PutTaskUnassign(c echo.Context) error {
	taskId, _ := strconv.Atoi(c.Param("id"))
	if err := data.UnassignTask(data.TaskId(taskId)); err != nil {
		log.Logger.Errorf("Error unassigning task: %v", err)
		return err
	}

	groups, err := data.GetGroups()
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "groups", groups)
}

func PutTask(c echo.Context) error {
	taskId, _ := strconv.Atoi(c.Param("id"))
	period, _ := strconv.Atoi(c.FormValue("period"))
	if err := data.UpdateTask(data.TaskId(taskId), data.Task{
		Name:        c.FormValue("name"),
		Description: c.FormValue("description"),
		Period:      period,
	}); err != nil {
		return err
	}
	groupId, _ := strconv.Atoi(c.FormValue("group"))
	if groupId != -1 {
		err := data.SetRelation(data.GroupId(groupId), data.TaskId(taskId))
		if err != nil {
			return err
		}
	}
	return c.String(http.StatusOK, "task modified successfully")
}

func GetEditTaskForm(c echo.Context) error {
	id, _ := strconv.Atoi(c.QueryParam("id"))
	task, err := data.GetTask(data.TaskId(id))
	if err != nil {
		log.Logger.Fatalf("GetEditTaskForm: %v", err)
	}

	return c.Render(http.StatusOK, "task-form", map[string]any{
		"Id":          id,
		"Name":        task.Name,
		"Description": task.Description,
		"Period":      task.Period,
	})
}

func GetTasksCount(c echo.Context) error {
	days, _ := strconv.Atoi(c.QueryParam("days"))
	count, err := data.TasksCount(days)
	if err != nil {
		return err
	}
	return c.String(http.StatusOK, fmt.Sprintf("%d", count))
}

func GetTaskGroupName(c echo.Context) error {
	taskId, _ := strconv.Atoi(c.Param("id"))
	group, err := data.GetTaskGroupName(data.TaskId(taskId))
	if err != nil {
		return err
	}
	return c.String(http.StatusOK, group)
}

func GetModalInactive(c echo.Context) error {
	return c.Render(http.StatusOK, "modal-inactive", nil)
}

func GetModalTaskInfo(c echo.Context) error {
	id, _ := strconv.Atoi(c.QueryParam("id"))
	task, err := data.GetTask(data.TaskId(id))
	if err != nil {
		log.Logger.Fatalf("GetModalTaskInfo: %v", err)
	}
	return c.Render(http.StatusOK, "modal-task-info", map[string]any{
		"Name": task.Name,
		"Info": task.Description,
	})
}
