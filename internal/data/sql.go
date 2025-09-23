package data

import (
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed init.sql
var schema string

type SqliteData struct {
	*sql.DB
	logger *log.Logger
}

// Init opens or creates a SQLite DB file.
// Example connStr: "./tasks.db"
func (sd *SqliteData) Init(connStr string, logger *log.Logger) {
	sd.logger = logger
	sd.logger.Debugf("opening database connection %q", connStr)

	// Open connection
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		sd.logger.Fatal(err)
	}
	sd.DB = db

	// Initialise schema
	if _, err := sd.DB.Exec(string(schema)); err != nil {
		sd.logger.Fatal(err)
	}
}

// AddTask inserts a task and returns the new id.
func (sd *SqliteData) AddTask(task *Task) (id TaskId) {
	res, err := sd.DB.Exec(
		`INSERT into tasks (name, description, period, last_completed) 
		VALUES (?, ?, ?, ?)`,
		task.Name,
		task.Description,
		task.Period,
		task.LastCompleted.UTC().Format(time.RFC3339),
	)
	if err != nil {
		sd.logger.Fatal(err)
	}

	lid, err := res.LastInsertId()
	if err != nil {
		sd.logger.Fatal(err)
	}

	return TaskId(lid)
}

// AddGroup inserts a group and returns the new id.
func (sd *SqliteData) AddGroup(group *Group) (id GroupId) {
	res, err := sd.DB.Exec(
		`INSERT into groups (name) 
		VALUES (?)`,
		group.Name,
	)
	if err != nil {
		sd.logger.Fatal(err)
	}

	lid, err := res.LastInsertId()
	if err != nil {
		sd.logger.Fatal(err)
	}

	return GroupId(lid)
}

// CompleteTask set a task as completed with the current timestamp.
func (sd *SqliteData) CompleteTask(id TaskId) error {
	_, err := sd.DB.Exec("UPDATE tasks SET last_completed=? WHERE id=?", time.Now().UTC().Format(time.RFC3339), id)
	return err
}

// SetRelation assign a list of tasks to the specified group.
// Both the group and the tasks are specified by their ids.
func (sd *SqliteData) SetRelation(groupId GroupId, taskIds ...TaskId) error {
	for _, taskId := range taskIds {
		if _, err := sd.DB.Exec("UPDATE tasks SET group_id=? WHERE id=?", groupId, taskId); err != nil {
			return err
		}
	}
	return nil
}

// GetTasksByGroup return all the tasks that are assigned to the specified group id.
// The task are returned as a list of pointer to the Task object.
func (sd *SqliteData) GetTasksByGroup(groupId GroupId) []*Task {
	rows, err := sd.DB.Query(
		`SELECT id, name, description, period, last_completed 
		FROM tasks 
		WHERE group_id=?`,
		groupId,
	)
	if err != nil {
		sd.logger.Fatal(err)
	}
	defer rows.Close()

	res, err := scanTasks(rows)
	if err != nil {
		sd.logger.Fatal(err)
	}
	return res
}

// GetUnassignedTasks return all the tasks that are not assigned to any group.
// The task are returned as a list of pointer to the Task object.
func (sd *SqliteData) GetUnassignedTasks() []*Task {
	rows, err := sd.DB.Query(
		`SELECT id, name, description, period, last_completed 
		FROM tasks 
		WHERE group_id is NULL`,
	)
	if err != nil {
		sd.logger.Fatal(err)
	}
	defer rows.Close()

	res, err := scanTasks(rows)
	if err != nil {
		sd.logger.Fatal(err)
	}
	return res
}

// GetGroups returns a list of pointers to all the groups.
func (sd *SqliteData) GetGroups() []*Group {
	rows, err := sd.DB.Query("SELECT id, name FROM groups")
	if err != nil {
		sd.logger.Fatal(err)
	}
	defer rows.Close()

	res, err := scanGroups(rows)
	if err != nil {
		sd.logger.Fatal(err)
	}
	return res
}

// GetTask retrieves a task by the specified id and returns a pointer to the parsed Task object.
func (sd *SqliteData) GetTask(id TaskId) *Task {
	var name, description, lastCompleted string
	var period int
	err := sd.DB.QueryRow(
		`SELECT name, description, period, last_completed 
		FROM tasks 
		WHERE id=?`,
		id,
	).Scan(&name, &description, &period, &lastCompleted)
	if err != nil {
		sd.logger.Fatal(err)
	}

	lastCompletedDate, _ := time.Parse(time.RFC3339, lastCompleted)
	return &Task{
		Id:            id,
		Name:          name,
		Description:   description,
		Period:        period,
		LastCompleted: lastCompletedDate,
	}
}

// UnassignTask removes the assigned group from the specified task.
// The task is specified by its id.
func (sd *SqliteData) UnassignTask(id TaskId) error {
	_, err := sd.DB.Exec(
		`UPDATE tasks 
		SET group_id=NULL 
		WHERE id=?`,
		id,
	)
	return err
}

// DeleteTask deletes the task specified by the id.
func (sd *SqliteData) DeleteTask(id TaskId) error {
	_, err := sd.DB.Exec(
		`DELETE FROM tasks 
		WHERE id=?`,
		id,
	)
	return err
}

// DeleteTask deletes the group specified by the id.
func (sd *SqliteData) DeleteGroup(id GroupId) error {
	_, err := sd.DB.Exec(
		`DELETE FROM groups 
		WHERE id=?`,
		id,
	)
	return err
}

// UpdateTask replace the task specified by the given id with the task provided by the given pointer.
func (sd *SqliteData) UpdateTask(id TaskId, task *Task) error {
	_, err := sd.DB.Exec(
		`UPDATE tasks 
		SET name=?, description=?, period=?
		WHERE id=?`,
		task.Name, task.Description, task.Period,
		id,
	)
	return err
}

// GetGroup retrieve the group specified by the id and returns the pointer to the parsed Group object.
func (sd *SqliteData) GetGroup(id GroupId) *Group {
	var name string
	err := sd.DB.QueryRow(
		`SELECT name 
		FROM groups 
		WHERE id=?`,
		id,
	).Scan(&name)
	if err != nil {
		sd.logger.Fatal(err)
	}

	return &Group{
		Id:   id,
		Name: name,
	}
}

// GetTaskGroupName retrieve the name of the group assigned to the specified task id.
func (sd *SqliteData) GetTaskGroupName(id TaskId) (string, error) {
	rows, err := sd.DB.Query(
		`SELECT g.name FROM groups g 
		JOIN tasks t ON g.id = t.group_id WHERE t.id=?`,
		id,
	)
	if err != nil {
		return "", fmt.Errorf("sql error while getting group name for task %d: %w", id, err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "no group", nil
	}

	var name string
	if err = rows.Scan(&name); err != nil {
		return "", fmt.Errorf("error scanning group name for task %d: %w", id, err)
	}
	return name, nil
}

// Tasks returns all the tasks filtered by the provided group id, days and expiration status.
func (sd *SqliteData) Tasks(groupId string, days string, expired bool) ([]*Task, error) {
	query := `SELECT id, name, description, period, last_completed FROM tasks`
	conds := make([]string, 0)
	args := make([]any, 0)

	if groupId != "" {
		if groupId != "-1" {
			conds = append(conds, "group_id = ?")
			args = append(args, groupId)
		} else {
			conds = append(conds, "group_id IS NULL")
		}
	}
	if days != "" {
		// <= DATE('now', '+' || ? || ' days')
		conds = append(conds, `DATE(last_completed, '+' || period || ' days') <= DATE('now', '+' || ? || ' days')`)
		args = append(args, days)
		if !expired {
			conds = append(conds, `DATE(last_completed, '+' || period || ' days') > DATE('now')`)
		}
	}

	if len(conds) > 0 {
		query += " WHERE " + joinAND(conds)
	}
	query += ` ORDER BY DATE(last_completed, '+' || period || ' days');`

	rows, err := sd.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("sql error while getting filtered tasks: %w", err)
	}
	defer rows.Close()

	res, err := scanTasks(rows)
	if err != nil {
		return nil, fmt.Errorf("sql error while getting filtered tasks: %w", err)
	}
	return res, nil
}

// TasksCount returns the amount of tasks due within the specified days.
func (sd *SqliteData) TasksCount(days int) (int, error) {
	var count int
	err := sd.DB.QueryRow(
		`SELECT COUNT(id) FROM tasks
         WHERE DATE(last_completed, '+' || period || ' days') <= DATE('now', '+' || ? || ' days')`,
		days,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// HELPERS

func scanTasks(rows *sql.Rows) ([]*Task, error) {
	var res []*Task
	for rows.Next() {
		var (
			id            TaskId
			name          string
			description   string
			period        int
			lastCompleted string
		)
		if err := rows.Scan(&id, &name, &description, &period, &lastCompleted); err != nil {
			return nil, err
		}
		dt, _ := time.Parse(time.RFC3339, lastCompleted)
		res = append(res, &Task{
			Id:            id,
			Name:          name,
			Description:   description,
			Period:        period,
			LastCompleted: dt,
		})
	}
	return res, nil
}

func scanGroups(rows *sql.Rows) ([]*Group, error) {
	groups := make([]*Group, 0)
	for rows.Next() {
		var id GroupId
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		groups = append(groups, &Group{
			Id:   id,
			Name: name,
		})
	}

	return groups, nil
}

func joinAND(xs []string) string {
	if len(xs) == 0 {
		return ""
	}
	out := xs[0]
	for i := 1; i < len(xs); i++ {
		out += " AND " + xs[i]
	}
	return out
}
