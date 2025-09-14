package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/markor147/peverel/internal/log"
	ts "github.com/markor147/peverel/internal/tasks"
)

func GetNewTaskForm(c echo.Context) error {
	return c.Render(http.StatusOK, "task-form", nil)
}

func GetNewGroupForm(c echo.Context) error {
	return c.Render(http.StatusOK, "new-group", nil)
}

func PostTask(c echo.Context) error {
	period, _ := strconv.Atoi(c.FormValue("period"))
	taskId := data.AddTask(&ts.Task{
		Name:        c.FormValue("name"),
		Description: c.FormValue("description"),
		Period:      period,
	})

	groupId, _ := strconv.Atoi(c.FormValue("group"))
	if groupId != -1 {
		err := data.SetRelation(ts.GroupId(groupId), taskId)
		if err != nil {
			log.Logger.Errorf("Error adding group %d to task %d: %v", groupId, taskId, err)
		}
	}

	return c.String(http.StatusOK, "task created successfully")
}

func PostGroup(c echo.Context) error {
	data.AddGroup(&ts.Group{
		Name: c.FormValue("name"),
	})
	return c.NoContent(http.StatusOK)
}

func GetTasks(c echo.Context) error {
	var tasks []*ts.Task
	groupId := c.QueryParam("group")
	days := c.QueryParam("days")
	expired := c.QueryParam("expired") != "false"

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
	case "levels":
		template = "tasks-levels"
	default:
		template = "tasks-levels"
	}

	return c.Render(http.StatusOK, template, tasks)
}

func GetGroups(c echo.Context) error {
	template := "groups"
	layout := c.QueryParam("layout")
	if layout == "options" {
		template = "groups-options"
	}
	return c.Render(http.StatusOK, template, data.GetGroups())
}

func PutTaskComplete(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	taskId := ts.TaskId(id)
	_ = data.CompleteTask(taskId)

	tasks, err := data.Tasks("", "", true)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "tasks-table", tasks)
}

func CreateMockTasks(c echo.Context) error {
	now := time.Now()

	bath := data.AddGroup(&ts.Group{
		Name: "Bathroom",
	})
	toilet := data.AddTask(&ts.Task{
		Name:          "Toilet",
		Description:   "Clean the toilet",
		Period:        2,
		LastCompleted: now,
	})
	bathroomFixtures := data.AddTask(&ts.Task{
		Name:          "Sink and shower",
		Description:   "Clean the sink and the shower",
		Period:        7,
		LastCompleted: now,
	})
	bathroomFloor := data.AddTask(&ts.Task{
		Name:          "Bathroom floor",
		Description:   "Vacuum and mop the bathroom floor",
		Period:        7,
		LastCompleted: now,
	})
	data.SetRelation(bath, toilet, bathroomFixtures, bathroomFloor)

	upHallway := data.AddGroup(&ts.Group{
		Name: "Up Hallway",
	})
	upHallwayCarpet := data.AddTask(&ts.Task{
		Name:          "Up Hallway carpet",
		Description:   "Vacuum the up hallway carpet",
		Period:        10,
		LastCompleted: now,
	})
	data.SetRelation(upHallway, upHallwayCarpet)

	studio := data.AddGroup(&ts.Group{
		Name: "Studio",
	})
	studioFloor := data.AddTask(&ts.Task{
		Name:          "Studio floor",
		Description:   "Vacuum and mop studio floor",
		Period:        10,
		LastCompleted: now,
	})
	studioDesk := data.AddTask(&ts.Task{
		Name:          "Studio desk",
		Description:   "Tidy up and dust the study desk",
		Period:        7,
		LastCompleted: now,
	})
	studioShelves := data.AddTask(&ts.Task{
		Name:          "Studio shelves",
		Description:   "Tidy up and dust the study shelves",
		Period:        14,
		LastCompleted: now,
	})
	data.SetRelation(studio, studioShelves, studioFloor, studioDesk)

	guest := data.AddGroup(&ts.Group{
		Name: "Guest Room",
	})
	guestFloor := data.AddTask(&ts.Task{
		Name:          "Guest room floor",
		Description:   "Vacuum the guest room carpet",
		Period:        30,
		LastCompleted: now,
	})
	guestFurniture := data.AddTask(&ts.Task{
		Name:          "Guest room furniture",
		Description:   "Tidy up and dust the guest room furniture",
		Period:        14,
		LastCompleted: now,
	})
	data.SetRelation(guest, guestFloor, guestFurniture)

	bed := data.AddGroup(&ts.Group{
		Name: "Bedroom",
	})
	bedFloor := data.AddTask(&ts.Task{
		Name:          "Guest room floor",
		Description:   "Vacuum the guest room carpet",
		Period:        7,
		LastCompleted: now,
	})
	bedFurniture := data.AddTask(&ts.Task{
		Name:          "Guest room furniture",
		Description:   "Tidy up and dust the guest room furniture",
		Period:        10,
		LastCompleted: now,
	})
	bedSheets := data.AddTask(&ts.Task{
		Name:          "Sheets",
		Description:   "Change the sheets in the bedroom",
		Period:        7,
		LastCompleted: now,
	})
	data.SetRelation(bed, bedSheets, bedFloor, bedFurniture)

	kitchen := data.AddGroup(&ts.Group{
		Name: "Kitchen",
	})
	kitchenSink := data.AddTask(&ts.Task{
		Name:          "Kitchen sink",
		Description:   "Clean and remove limescale from the kitchen sink",
		Period:        3,
		LastCompleted: now,
	})
	kitchenFloor := data.AddTask(&ts.Task{
		Name:          "Kitchen floor",
		Description:   "Vacuum and mop the kitchen floor",
		Period:        7,
		LastCompleted: now,
	})
	fridge := data.AddTask(&ts.Task{
		Name:          "Fridge",
		Description:   "Clean the fridge",
		Period:        30,
		LastCompleted: now,
	})
	kitchenSurfaces := data.AddTask(&ts.Task{
		Name:          "Kitchen surfaces",
		Description:   "Clean the surfaces in the kitchen",
		Period:        7,
		LastCompleted: now,
	})
	kitchenTidy := data.AddTask(&ts.Task{
		Name:          "Kitchen items",
		Description:   "Tidy up the items in the kitchen cabinets",
		Period:        30,
		LastCompleted: now,
	})
	data.SetRelation(kitchen, kitchenSurfaces, kitchenTidy, kitchenSink, kitchenFloor, fridge)

	living := data.AddGroup(&ts.Group{
		Name: "Living Room",
	})
	livingFloor := data.AddTask(&ts.Task{
		Name:          "Living room carpet",
		Description:   "Vacuum the living room carpet",
		Period:        10,
		LastCompleted: now,
	})
	livingFurnitures := data.AddTask(&ts.Task{
		Name:          "Living room furniture",
		Description:   "Tidy up and dust the living room furniture",
		Period:        7,
		LastCompleted: now,
	})
	sofa := data.AddTask(&ts.Task{
		Name:          "Sofa",
		Description:   "Change the sheet of the sofa",
		Period:        14,
		LastCompleted: now,
	})
	data.SetRelation(living, livingFurnitures, livingFloor, sofa)

	hall := data.AddGroup(&ts.Group{
		Name: "Hall",
	})
	hallFloor := data.AddTask(&ts.Task{
		Name:          "Hall floor",
		Description:   "Vacuum and mop the hall floor",
		Period:        7,
		LastCompleted: now,
	})
	shoeRack := data.AddTask(&ts.Task{
		Name:          "Shoerack",
		Description:   "Clean the shoe rack",
		Period:        30,
		LastCompleted: now,
	})
	data.SetRelation(hall, hallFloor, shoeRack)

	stairs := data.AddGroup(&ts.Group{
		Name: "Stairs",
	})
	stairsCarpet := data.AddTask(&ts.Task{
		Name:          "Stairs carpet",
		Description:   "Vacuum the stairs carpet",
		Period:        14,
		LastCompleted: now,
	})
	handrail := data.AddTask(&ts.Task{
		Name:          "Handrail",
		Description:   "Clean the stairs handrail",
		Period:        30,
		LastCompleted: now,
	})
	data.SetRelation(stairs, stairsCarpet, handrail)

	downHallway := data.AddGroup(&ts.Group{
		Name: "Down Hallway",
	})
	downHallwayFloor := data.AddTask(&ts.Task{
		Name:          "Down hallway floor",
		Description:   "Vacuum and mop the down hallway floor",
		Period:        7,
		LastCompleted: now,
	})
	data.SetRelation(downHallway, downHallwayFloor)

	data.AddTask(&ts.Task{
		Name:          "Marco's bike cleaning",
		Description:   "Clean frame and chain",
		Period:        30,
		LastCompleted: now,
	})
	data.AddTask(&ts.Task{
		Name:          "Marzia's bike cleaning",
		Description:   "Clean frame and chain",
		Period:        30,
		LastCompleted: now,
	})
	data.AddTask(&ts.Task{
		Name:          "Marco's bike oil",
		Description:   "Put oil on chain and gears",
		Period:        7,
		LastCompleted: now,
	})
	data.AddTask(&ts.Task{
		Name:          "Marzia's bike oil",
		Description:   "Put oil on chain and gears",
		Period:        7,
		LastCompleted: now,
	})

	return c.String(http.StatusOK, "mock tasks created")
}

func GetTaskNextTime(c echo.Context) error {
	taskId, _ := strconv.Atoi(c.Param("id"))
	return c.HTML(http.StatusOK, renderTaskNextTime(ts.TaskId(taskId)))
}

func renderTaskNextTime(taskId ts.TaskId) string {
	task := data.GetTask(taskId)

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

	err := data.SetRelation(ts.GroupId(groupId), ts.TaskId(taskId))
	if err != nil {
		log.Logger.Errorf("add group assign task err: %v", err)
	}

	return c.Render(http.StatusOK, "groups", data.GetGroups())
}

func DeleteTask(c echo.Context) error {
	taskId, _ := strconv.Atoi(c.Param("id"))
	if err := data.DeleteTask(ts.TaskId(taskId)); err != nil {
		log.Logger.Errorf("Error deleting task: %v", err)
		return err
	}
	return c.String(http.StatusOK, "task deleted successfully")
}

func DeleteGroup(c echo.Context) error {
	groupId, _ := strconv.Atoi(c.Param("id"))
	if err := data.DeleteGroup(ts.GroupId(groupId)); err != nil {
		log.Logger.Errorf("Error deleting group: %v", err)
		return err
	}
	return c.Render(http.StatusOK, "groups", data.GetGroups())
}

func PutTaskUnassign(c echo.Context) error {
	taskId, _ := strconv.Atoi(c.Param("id"))
	if err := data.UnassignTask(ts.TaskId(taskId)); err != nil {
		log.Logger.Errorf("Error unassigning task: %v", err)
		return err
	}
	return c.Render(http.StatusOK, "groups", data.GetGroups())
}

func PutTask(c echo.Context) error {
	taskId, _ := strconv.Atoi(c.Param("id"))
	period, _ := strconv.Atoi(c.FormValue("period"))
	if err := data.UpdateTask(ts.TaskId(taskId), &ts.Task{
		Name:        c.FormValue("name"),
		Description: c.FormValue("description"),
		Period:      period,
	}); err != nil {
		return err
	}
	groupId, _ := strconv.Atoi(c.FormValue("group"))
	if groupId != -1 {
		err := data.SetRelation(ts.GroupId(groupId), ts.TaskId(taskId))
		if err != nil {
			log.Logger.Errorf("Error adding group %d to task %d: %v", groupId, taskId, err)
		}
	}
	return c.String(http.StatusOK, "task modified successfully")
}

func GetEditTaskForm(c echo.Context) error {
	id, _ := strconv.Atoi(c.QueryParam("id"))
	task := data.GetTask(ts.TaskId(id))
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
	group, err := data.GetTaskGroupName(ts.TaskId(taskId))
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
	task := data.GetTask(ts.TaskId(id))
	return c.Render(http.StatusOK, "modal-task-info", map[string]any{
		"Name": task.Name,
		"Info": task.Description,
	})
}
