package data

import (
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	"github.com/markor147/peverel/internal/log"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed init.sql
var schema string

var db *sql.DB

// Init opens or creates a SQLite DB file.
// Example connStr: "./tasks.db"
func Init(connStr string) error {
	log.Logger.Debugf("opening database connection %q", connStr)

	// Open connection
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return fmt.Errorf("init db: %w", err)
	}

	// Initialise schema
	if _, err := db.Exec(string(schema)); err != nil {
		return fmt.Errorf("init db: %w", err)
	}

	return nil
}

// AddTask inserts a task and returns the new id.
func AddTask(task Task) (TaskId, error) {
	res, err := db.Exec(
		`INSERT into tasks (name, description, period, last_completed) 
		VALUES (?, ?, ?, ?)`,
		task.Name,
		task.Description,
		task.Period,
		task.LastCompleted.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return -1, fmt.Errorf("function AddTask: %w", err)
	}

	lid, err := res.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("function AddTask: %w", err)
	}

	return TaskId(lid), nil
}

// AddGroup inserts a group and returns the new id.
func AddGroup(group *Group) (GroupId, error) {
	res, err := db.Exec(
		`INSERT into groups (name) 
		VALUES (?)`,
		group.Name,
	)
	if err != nil {
		return -1, fmt.Errorf("function AddGroup: %w", err)
	}

	lid, err := res.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("function AddGroup: %w", err)
	}

	return GroupId(lid), nil
}

// CompleteTask set a task as completed with the current timestamp.
func CompleteTask(id TaskId) error {
	_, err := db.Exec("UPDATE tasks SET last_completed=? WHERE id=?", time.Now().UTC().Format(time.RFC3339), id)
	return err
}

// SetRelation assign a list of tasks to the specified group.
// Both the group and the tasks are specified by their ids.
func SetRelation(groupId GroupId, taskIds ...TaskId) error {
	for _, taskId := range taskIds {
		if _, err := db.Exec("UPDATE tasks SET group_id=? WHERE id=?", groupId, taskId); err != nil {
			return fmt.Errorf("assign group %d to task %d: %w", groupId, taskId, err)
		}
	}
	return nil
}

// GetGroups returns a list of pointers to all the groups.
func GetGroups() ([]Group, error) {
	rows, err := db.Query("SELECT id, name FROM groups")
	if err != nil {
		return nil, fmt.Errorf("function GetGroups: %w", err)
	}
	defer rows.Close()

	res, err := scanGroups(rows)
	if err != nil {
		return nil, fmt.Errorf("function GetGroups: %w", err)
	}
	return res, nil
}

// GetTask retrieves a task by the specified id and returns a pointer to the parsed Task object.
func GetTask(id TaskId) (Task, error) {
	var name, description, lastCompleted string
	var period int
	err := db.QueryRow(
		`SELECT name, description, period, last_completed 
		FROM tasks 
		WHERE id=?`,
		id,
	).Scan(&name, &description, &period, &lastCompleted)
	if err != nil {
		return Task{}, fmt.Errorf("function GetTask: %w", err)
	}

	lastCompletedDate, _ := time.Parse(time.RFC3339, lastCompleted)
	return Task{
		Id:            id,
		Name:          name,
		Description:   description,
		Period:        period,
		LastCompleted: lastCompletedDate,
	}, nil
}

// UnassignTask removes the assigned group from the specified task.
// The task is specified by its id.
func UnassignTask(id TaskId) error {
	_, err := db.Exec(
		`UPDATE tasks 
		SET group_id=NULL 
		WHERE id=?`,
		id,
	)
	return err
}

// DeleteTask deletes the task specified by the id.
func DeleteTask(id TaskId) error {
	_, err := db.Exec(
		`DELETE FROM tasks 
		WHERE id=?`,
		id,
	)
	return err
}

// DeleteTask deletes the group specified by the id.
func DeleteGroup(id GroupId) error {
	_, err := db.Exec(
		`DELETE FROM groups 
		WHERE id=?`,
		id,
	)
	return err
}

// UpdateTask replace the task specified by the given id with the task provided by the given pointer.
func UpdateTask(id TaskId, task Task) error {
	_, err := db.Exec(
		`UPDATE tasks 
		SET name=?, description=?, period=?
		WHERE id=?`,
		task.Name, task.Description, task.Period,
		id,
	)
	return err
}

// GetTaskGroupName retrieve the name of the group assigned to the specified task id.
func GetTaskGroupName(id TaskId) (string, error) {
	rows, err := db.Query(
		`SELECT g.name FROM groups g 
		JOIN tasks t ON g.id = t.group_id WHERE t.id=?`,
		id,
	)
	if err != nil {
		return "", fmt.Errorf("get group name for task %d: %w", id, err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "no group", nil
	}

	var name string
	if err = rows.Scan(&name); err != nil {
		return "", fmt.Errorf("parse group name for task %d: %w", id, err)
	}
	return name, nil
}

// Tasks returns all the tasks filtered by the provided group id, days and expiration status.
func Tasks(groupId string, days string, expired bool) ([]Task, error) {
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

	rows, err := db.Query(query, args...)
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
func TasksCount(days int) (int, error) {
	var count int
	err := db.QueryRow(
		`SELECT COUNT(id) FROM tasks
         WHERE DATE(last_completed, '+' || period || ' days') <= DATE('now', '+' || ? || ' days')`,
		days,
	).Scan(&count)
	if err != nil {
		return -1, err
	}
	return count, nil
}

/* === Unusued === */

// GetTasksByGroup return all the tasks that are assigned to the specified group id.
// The task are returned as a list of pointer to the Task object.
func GetTasksByGroup(groupId GroupId) ([]Task, error) {
	rows, err := db.Query(
		`SELECT id, name, description, period, last_completed
		FROM tasks
		WHERE group_id=?`,
		groupId,
	)
	if err != nil {
		return nil, fmt.Errorf("function GetTasksByGroup: %w", err)
	}
	defer rows.Close()

	res, err := scanTasks(rows)
	if err != nil {
		return nil, fmt.Errorf("function GetTasksByGroup: %w", err)
	}
	return res, nil
}

// GetUnassignedTasks return all the tasks that are not assigned to any group.
// The task are returned as a list of pointer to the Task object.
func GetUnassignedTasks() ([]Task, error) {
	rows, err := db.Query(
		`SELECT id, name, description, period, last_completed
		FROM tasks
		WHERE group_id is NULL`,
	)
	if err != nil {
		return nil, fmt.Errorf("function GetUnassignedTasks: %w", err)
	}
	defer rows.Close()

	res, err := scanTasks(rows)
	if err != nil {
		return nil, fmt.Errorf("function GetUnassignedTasks: %w", err)
	}
	return res, nil
}

// GetGroup retrieve the group specified by the id and returns the pointer to the parsed Group object.
func GetGroup(id GroupId) (Group, error) {
	var name string
	err := db.QueryRow(
		`SELECT name
		FROM groups
		WHERE id=?`,
		id,
	).Scan(&name)
	if err != nil {
		return Group{}, fmt.Errorf("function GetGroup: %w", err)
	}

	return Group{
		Id:   id,
		Name: name,
	}, nil
}

/* === Helpers === */

func scanTasks(rows *sql.Rows) ([]Task, error) {
	res := make([]Task, 0)
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
		res = append(res, Task{
			Id:            id,
			Name:          name,
			Description:   description,
			Period:        period,
			LastCompleted: dt,
		})
	}
	return res, nil
}

func scanGroups(rows *sql.Rows) ([]Group, error) {
	res := make([]Group, 0)
	for rows.Next() {
		var id GroupId
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		res = append(res, Group{
			Id:   id,
			Name: name,
		})
	}

	return res, nil
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
