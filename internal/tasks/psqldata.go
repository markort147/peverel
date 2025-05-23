package tasks

import (
	"database/sql"
	"fmt"
	"github.com/labstack/gommon/log"
	_ "github.com/lib/pq"
	"strings"
	"time"
)

type PsqlData struct {
	*sql.DB
	logger *log.Logger
}

func (pd *PsqlData) Init(connStr string, logger *log.Logger) {
	pd.logger = logger
	pd.logger.Debugf("opening database connection %q", connStr)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		pd.logger.Fatal(err)
	}
	pd.DB = db
}

func (pd *PsqlData) AddTask(task *Task) (id TaskId) {
	err := pd.DB.QueryRow("INSERT into tasks (name, description, period, last_completed) VALUES ($1, $2, $3, $4) RETURNING id",
		task.Name,
		task.Description,
		task.Period,
		task.LastCompleted,
	).Scan(&id)
	if err != nil {
		pd.logger.Fatal(err)
	}
	return id
}

func (pd *PsqlData) AddGroup(group *Group) (id GroupId) {
	err := pd.DB.QueryRow("INSERT into groups (name) VALUES ($1) RETURNING id",
		group.Name,
	).Scan(&id)
	//defer rows.Close()
	if err != nil {
		pd.logger.Fatal(err)
	}
	return id
}

func (pd *PsqlData) CompleteTask(id TaskId) error {
	_, err := pd.DB.Exec("UPDATE tasks SET last_completed=$1 WHERE id=$2", time.Now(), id)
	return err
}

func (pd *PsqlData) SetRelation(groupId GroupId, taskIds ...TaskId) error {
	for _, taskId := range taskIds {
		_, err := pd.DB.Exec("UPDATE tasks SET group_id=$1 WHERE id=$2", groupId, taskId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pd *PsqlData) GetTasksByGroup(groupId GroupId) []*Task {
	rows, err := pd.DB.Query("SELECT id, name, description, period, last_completed FROM tasks where group_id=$1", groupId)
	defer rows.Close()
	if err != nil {
		pd.logger.Fatal(err)
	}

	res := make([]*Task, 0)
	for rows.Next() {
		var id TaskId
		var name string
		var description string
		var period int
		var lastCompleted string
		rows.Scan(&id, &name, &description, &period, &lastCompleted)
		lastCompletedDate, _ := time.Parse("2006-01-02T15:04:05Z", lastCompleted)
		res = append(res, &Task{
			Id:            id,
			Name:          name,
			Description:   description,
			Period:        period,
			LastCompleted: lastCompletedDate,
		})
	}

	return res
}

func (pd *PsqlData) GetUnassignedTasks() []*Task {
	rows, err := pd.DB.Query("SELECT id, name, description, period, last_completed FROM tasks where group_id is NULL")
	defer rows.Close()
	if err != nil {
		pd.logger.Fatal(err)
	}

	unassignedTasks := make([]*Task, 0)
	for rows.Next() {
		var id TaskId
		var name string
		var description string
		var period int
		var lastCompleted string
		rows.Scan(&id, &name, &description, &period, &lastCompleted)
		lastCompletedDate, _ := time.Parse("2006-01-02T15:04:05Z", lastCompleted)
		unassignedTasks = append(unassignedTasks, &Task{
			Id:            id,
			Name:          name,
			Description:   description,
			Period:        period,
			LastCompleted: lastCompletedDate,
		})
	}

	return unassignedTasks
}

func (pd *PsqlData) GetGroups() []*Group {
	rows, err := pd.DB.Query("SELECT id, name from groups")
	defer rows.Close()
	if err != nil {
		pd.logger.Fatal(err)
	}

	groups := make([]*Group, 0)
	for rows.Next() {
		var id GroupId
		var name string
		rows.Scan(&id, &name)
		groups = append(groups, &Group{
			Id:   id,
			Name: name,
		})
	}

	return groups
}

func (pd *PsqlData) GetTask(id TaskId) *Task {
	var name string
	var description string
	var period int
	var lastCompleted string
	err := pd.DB.QueryRow("SELECT name, description, period, last_completed FROM tasks WHERE id=$1", id).Scan(&name, &description, &period, &lastCompleted)
	//defer rows.Close()
	if err != nil {
		pd.logger.Fatal(err)
	}

	//rows.Scan(&name, &description, &period, &lastCompleted)
	lastCompletedDate, _ := time.Parse("2006-01-02T15:04:05Z", lastCompleted)
	return &Task{
		Id:            id,
		Name:          name,
		Description:   description,
		Period:        period,
		LastCompleted: lastCompletedDate,
	}
}

func (pd *PsqlData) UnassignTask(id TaskId) error {
	_, err := pd.DB.Exec("UPDATE tasks SET group_id=NULL WHERE id=$1", id)
	if err != nil {
		return err
	}
	return nil
}

func (pd *PsqlData) DeleteTask(id TaskId) error {
	_, err := pd.DB.Exec("DELETE FROM tasks WHERE id=$1", id)
	if err != nil {
		return err
	}
	return nil
}

func (pd *PsqlData) DeleteGroup(id GroupId) error {
	_, err := pd.DB.Exec("DELETE FROM groups WHERE id=$1", id)
	if err != nil {
		return err
	}
	return nil
}

func (pd *PsqlData) UpdateTask(id TaskId, task *Task) error {
	_, err := pd.DB.Exec("UPDATE tasks SET name=$1, description=$2, period=$3 WHERE id=$4",
		task.Name,
		task.Description,
		task.Period,
		id,
	)
	if err != nil {
		return err
	}
	return nil
}

func (pd *PsqlData) GetGroup(id GroupId) *Group {
	var name string
	err := pd.DB.QueryRow("SELECT name FROM groups WHERE id=$1", id).Scan(&name)
	if err != nil {
		pd.logger.Fatal(err)
	}
	return &Group{
		Id:   id,
		Name: name,
	}
}

func (pd *PsqlData) GetTaskGroupName(id TaskId) (string, error) {
	var name string
	rows, err := pd.DB.Query("SELECT g.name FROM groups g JOIN tasks t ON g.id = t.group_id WHERE t.id=$1", id)
	if err != nil {
		return "", fmt.Errorf("sql error while getting group name for task %d: %w", id, err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "no group", nil
	}

	if err = rows.Scan(&name); err != nil {
		return "", fmt.Errorf("error scanning group name for task %d: %w", id, err)
	}
	return name, nil
}

func (pd *PsqlData) Tasks(groupId string, days string, expired bool) ([]*Task, error) {
	query := "SELECT id, name, description, period, last_completed FROM tasks"

	conditions := make([]string, 0)
	if groupId != "" {
		if groupId != "-1" {
			conditions = append(conditions, fmt.Sprintf("group_id = %s", groupId))
		} else {
			conditions = append(conditions, "group_id IS NULL")
		}
	}
	if days != "" {
		conditions = append(conditions, fmt.Sprintf("DATE_TRUNC('day', last_completed + period * INTERVAL '1 day') <= DATE_TRUNC('day', CURRENT_DATE + interval '%s days')", days))
		if !expired {
			conditions = append(conditions, "DATE_TRUNC('day', last_completed + period * INTERVAL '1 day') > DATE_TRUNC('day', CURRENT_DATE)")
		}
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY last_completed + period * INTERVAL '1 day';"

	rows, err := pd.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("sql error while getting filtered tasks: %w", err)
	}
	defer rows.Close()

	res := make([]*Task, 0)
	for rows.Next() {
		var id TaskId
		var name string
		var description string
		var period int
		var lastCompleted string
		rows.Scan(&id, &name, &description, &period, &lastCompleted)
		lastCompletedDate, _ := time.Parse("2006-01-02T15:04:05Z", lastCompleted)
		pd.logger.Debugf("LastCompleted: %s - LastCompletedDate: %v", lastCompleted, lastCompletedDate)
		res = append(res, &Task{
			Id:            id,
			Name:          name,
			Description:   description,
			Period:        period,
			LastCompleted: lastCompletedDate,
		})
	}

	return res, nil
}

func (pd *PsqlData) TasksCount(days int) (int, error) {
	var count int
	err := pd.DB.QueryRow(`SELECT COUNT(id) FROM tasks 
WHERE DATE_TRUNC('day', last_completed + period * INTERVAL '1 day') <= DATE_TRUNC('day', CURRENT_DATE + $1 * interval '1 day')`, days).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
