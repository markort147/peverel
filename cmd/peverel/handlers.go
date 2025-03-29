package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
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

func GetTasks(c echo.Context) error {
	tasks := make(map[TaskId]*Task)
	groupId := c.QueryParam("group")
	switch groupId {
	case "":
		tasks = data.GetTasks()
	case "none":
		tasks = data.GetUnassignedTasks()
	default:
		intGroupId, _ := strconv.Atoi(groupId)
		tasks = data.GetTasksByGroup(GroupId(intGroupId))
	}

	template := "tasks-table"
	layout := c.QueryParam("layout")
	if layout == "options" {
		template = "tasks-options"
	}

	return c.Render(http.StatusOK, template, tasks)
}

func GetGroups(c echo.Context) error {
	return c.Render(http.StatusOK, "groups", data.GetGroups())
}

func PutTaskComplete(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	taskId := TaskId(id)
	_ = data.CompleteTask(taskId)
	return c.HTML(http.StatusOK, renderTaskNextTime(taskId))
}

func CreateMockTasks(c echo.Context) error {
	now := time.Now()

	bath := data.AddGroup(&Group{
		Name: "Bathroom",
	})
	toilet := data.AddTask(&Task{
		Name:          "Toilet",
		Description:   "Clean the toilet",
		Period:        2,
		LastCompleted: now,
	})
	bathroomFixtures := data.AddTask(&Task{
		Name:          "Sink and shower",
		Description:   "Clean the sink and the shower",
		Period:        7,
		LastCompleted: now,
	})
	bathroomFloor := data.AddTask(&Task{
		Name:          "Bathroom floor",
		Description:   "Vacuum and mop the bathroom floor",
		Period:        7,
		LastCompleted: now,
	})
	data.AddRelation(bath, toilet, bathroomFixtures, bathroomFloor)

	upHallway := data.AddGroup(&Group{
		Name: "Up Hallway",
	})
	upHallwayCarpet := data.AddTask(&Task{
		Name:          "Up Hallway carpet",
		Description:   "Vacuum the up hallway carpet",
		Period:        10,
		LastCompleted: now,
	})
	data.AddRelation(upHallway, upHallwayCarpet)

	studio := data.AddGroup(&Group{
		Name: "Studio",
	})
	studioFloor := data.AddTask(&Task{
		Name:          "Studio floor",
		Description:   "Vacuum and mop studio floor",
		Period:        10,
		LastCompleted: now,
	})
	studioDesk := data.AddTask(&Task{
		Name:          "Studio desk",
		Description:   "Tidy up and dust the study desk",
		Period:        7,
		LastCompleted: now,
	})
	studioShelves := data.AddTask(&Task{
		Name:          "Studio shelves",
		Description:   "Tidy up and dust the study shelves",
		Period:        14,
		LastCompleted: now,
	})
	data.AddRelation(studio, studioShelves, studioFloor, studioDesk)

	guest := data.AddGroup(&Group{
		Name: "Guest Room",
	})
	guestFloor := data.AddTask(&Task{
		Name:          "Guest room floor",
		Description:   "Vacuum the guest room carpet",
		Period:        30,
		LastCompleted: now,
	})
	guestFurniture := data.AddTask(&Task{
		Name:          "Guest room furniture",
		Description:   "Tidy up and dust the guest room furniture",
		Period:        14,
		LastCompleted: now,
	})
	data.AddRelation(guest, guestFloor, guestFurniture)

	bed := data.AddGroup(&Group{
		Name: "Bedroom",
	})
	bedFloor := data.AddTask(&Task{
		Name:          "Guest room floor",
		Description:   "Vacuum the guest room carpet",
		Period:        7,
		LastCompleted: now,
	})
	bedFurniture := data.AddTask(&Task{
		Name:          "Guest room furniture",
		Description:   "Tidy up and dust the guest room furniture",
		Period:        10,
		LastCompleted: now,
	})
	bedSheets := data.AddTask(&Task{
		Name:          "Sheets",
		Description:   "Change the sheets in the bedroom",
		Period:        7,
		LastCompleted: now,
	})
	data.AddRelation(bed, bedSheets, bedFloor, bedFurniture)

	kitchen := data.AddGroup(&Group{
		Name: "Kitchen",
	})
	kitchenSink := data.AddTask(&Task{
		Name:          "Kitchen sink",
		Description:   "Clean and remove limescale from the kitchen sink",
		Period:        3,
		LastCompleted: now,
	})
	kitchenFloor := data.AddTask(&Task{
		Name:          "Kitchen floor",
		Description:   "Vacuum and mop the kitchen floor",
		Period:        7,
		LastCompleted: now,
	})
	fridge := data.AddTask(&Task{
		Name:          "Fridge",
		Description:   "Clean the fridge",
		Period:        30,
		LastCompleted: now,
	})
	kitchenSurfaces := data.AddTask(&Task{
		Name:          "Kitchen surfaces",
		Description:   "Clean the surfaces in the kitchen",
		Period:        7,
		LastCompleted: now,
	})
	kitchenTidy := data.AddTask(&Task{
		Name:          "Kitchen items",
		Description:   "Tidy up the items in the kitchen cabinets",
		Period:        30,
		LastCompleted: now,
	})
	data.AddRelation(kitchen, kitchenSurfaces, kitchenTidy, kitchenSink, kitchenFloor, fridge)

	living := data.AddGroup(&Group{
		Name: "Living Room",
	})
	livingFloor := data.AddTask(&Task{
		Name:          "Living room carpet",
		Description:   "Vacuum the living room carpet",
		Period:        10,
		LastCompleted: now,
	})
	livingFurnitures := data.AddTask(&Task{
		Name:          "Living room furniture",
		Description:   "Tidy up and dust the living room furniture",
		Period:        7,
		LastCompleted: now,
	})
	sofa := data.AddTask(&Task{
		Name:          "Sofa",
		Description:   "Change the sheet of the sofa",
		Period:        14,
		LastCompleted: now,
	})
	data.AddRelation(living, livingFurnitures, livingFloor, sofa)

	hall := data.AddGroup(&Group{
		Name: "Hall",
	})
	hallFloor := data.AddTask(&Task{
		Name:          "Hall floor",
		Description:   "Vacuum and mop the hall floor",
		Period:        7,
		LastCompleted: now,
	})
	shoeRack := data.AddTask(&Task{
		Name:          "Shoerack",
		Description:   "Clean the shoe rack",
		Period:        30,
		LastCompleted: now,
	})
	data.AddRelation(hall, hallFloor, shoeRack)

	stairs := data.AddGroup(&Group{
		Name: "Stairs",
	})
	stairsCarpet := data.AddTask(&Task{
		Name:          "Stairs carpet",
		Description:   "Vacuum the stairs carpet",
		Period:        14,
		LastCompleted: now,
	})
	handrail := data.AddTask(&Task{
		Name:          "Handrail",
		Description:   "Clean the stairs handrail",
		Period:        30,
		LastCompleted: now,
	})
	data.AddRelation(stairs, stairsCarpet, handrail)

	downHallway := data.AddGroup(&Group{
		Name: "Down Hallway",
	})
	downHallwayFloor := data.AddTask(&Task{
		Name:          "Down hallway floor",
		Description:   "Vacuum and mop the down hallway floor",
		Period:        7,
		LastCompleted: now,
	})
	data.AddRelation(downHallway, downHallwayFloor)

	data.AddTask(&Task{
		Name:          "Marco's bike cleaning",
		Description:   "Clean frame and chain",
		Period:        30,
		LastCompleted: now,
	})
	data.AddTask(&Task{
		Name:          "Marzia's bike cleaning",
		Description:   "Clean frame and chain",
		Period:        30,
		LastCompleted: now,
	})
	data.AddTask(&Task{
		Name:          "Marco's bike oil",
		Description:   "Put oil on chain and gears",
		Period:        7,
		LastCompleted: now,
	})
	data.AddTask(&Task{
		Name:          "Marzia's bike oil",
		Description:   "Put oil on chain and gears",
		Period:        7,
		LastCompleted: now,
	})

	return GetGroups(c)
}

func GetTaskNextTime(c echo.Context) error {
	taskId, _ := strconv.Atoi(c.Param("id"))
	return c.HTML(http.StatusOK, renderTaskNextTime(TaskId(taskId)))
}

func renderTaskNextTime(taskId TaskId) string {
	task := data.GetTask(taskId)

	nextDay := task.LastCompleted.AddDate(0, 0, task.Period)

	layout := "20060102"
	todayStr := time.Now().Format(layout)
	nextDayStr := nextDay.Format(layout)
	Logger.Debugf("last: %v - period: %d - next day: %v - string: %s", task.LastCompleted, task.Period, nextDay, nextDayStr)

	if todayStr == nextDayStr {
		return "today"
	}
	todayTime, _ := time.Parse(layout, todayStr)
	nextDayTime, _ := time.Parse(layout, nextDayStr)
	diff := int(nextDayTime.Sub(todayTime).Hours() / 24)
	Logger.Debugf("diff: %d", diff)
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
		Logger.Errorf("add group assign task err: %v", err)
	}

	return c.Render(http.StatusOK, "groups", data.GetGroups())
}
