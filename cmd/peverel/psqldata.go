package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"time"
)

type PsqlData struct {
	*sql.DB
}

func (pd *PsqlData) Init(connStr string) {
	Logger.Debugf("opening database connection %q", connStr)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		Logger.Fatal(err)
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
		Logger.Fatal(err)
	}
	//defer rows.Close()
	//if err != nil {
	//	Logger.Fatal(err)
	//}

	//last, _ := rows.LastInsertId()
	//id = TaskId(last)
	//rows.Scan(&id)
	return id
}

func (pd *PsqlData) AddGroup(group *Group) (id GroupId) {
	err := pd.DB.QueryRow("INSERT into groups (name) VALUES ($1) RETURNING id",
		group.Name,
	).Scan(&id)
	//defer rows.Close()
	if err != nil {
		Logger.Fatal(err)
	}

	//last, _ := rows.LastInsertId()
	//id = GroupId(last)
	//rows.Scan(&id)
	return id
}

func (pd *PsqlData) CompleteTask(id TaskId) error {
	_, err := pd.DB.Exec("UPDATE tasks SET last_completed=$1 WHERE id=$2", time.Now(), id)
	return err
}

func (pd *PsqlData) SetRelation(groupId GroupId, taskIds ...TaskId) error {
	for _, taskId := range taskIds {
		Logger.Debugf("adding task %d to group %d", taskId, groupId)
		_, err := pd.DB.Exec("UPDATE tasks SET group_id=$1 WHERE id=$2", groupId, taskId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pd *PsqlData) GetTasksByGroup(groupId GroupId) map[TaskId]*Task {
	rows, err := pd.DB.Query("SELECT id, name, description, period, last_completed FROM tasks where group_id=$1", groupId)
	defer rows.Close()
	if err != nil {
		Logger.Fatal(err)
	}

	res := make(map[TaskId]*Task)
	for rows.Next() {
		var id TaskId
		var name string
		var description string
		var period int
		var lastCompleted string
		rows.Scan(&id, &name, &description, &period, &lastCompleted)
		lastCompletedDate, _ := time.Parse("2006-01-02T15:04:05Z", lastCompleted)
		res[id] = &Task{
			Name:          name,
			Description:   description,
			Period:        period,
			LastCompleted: lastCompletedDate,
		}
	}

	return res
}

func (pd *PsqlData) GetUnassignedTasks() map[TaskId]*Task {
	rows, err := pd.DB.Query("SELECT id, name, description, period, last_completed FROM tasks where group_id is NULL")
	defer rows.Close()
	if err != nil {
		Logger.Fatal(err)
	}

	unassignedTasks := make(map[TaskId]*Task)
	for rows.Next() {
		var id TaskId
		var name string
		var description string
		var period int
		var lastCompleted string
		rows.Scan(&id, &name, &description, &period, &lastCompleted)
		lastCompletedDate, _ := time.Parse("2006-01-02T15:04:05Z", lastCompleted)
		unassignedTasks[id] = &Task{
			Name:          name,
			Description:   description,
			Period:        period,
			LastCompleted: lastCompletedDate,
		}
	}

	return unassignedTasks
}

func (pd *PsqlData) GetTasks() map[TaskId]*Task {
	rows, err := pd.DB.Query("SELECT id, name, description, period, last_completed FROM tasks")
	defer rows.Close()
	if err != nil {
		Logger.Fatal(err)
	}

	tasks := make(map[TaskId]*Task)
	for rows.Next() {
		var id TaskId
		var name string
		var description string
		var period int
		var lastCompleted string
		rows.Scan(&id, &name, &description, &period, &lastCompleted)
		lastCompletedDate, _ := time.Parse("2006-01-02T15:04:05Z", lastCompleted)
		Logger.Debugf("LastCompleted: %s - LastCompletedDate: %v", lastCompleted, lastCompletedDate)
		tasks[id] = &Task{
			Name:          name,
			Description:   description,
			Period:        period,
			LastCompleted: lastCompletedDate,
		}
		Logger.Debugf("Found task %+v", tasks[id])
	}

	return tasks
}

func (pd *PsqlData) GetGroups() map[GroupId]*Group {
	rows, err := pd.DB.Query("SELECT id, name from groups")
	defer rows.Close()
	if err != nil {
		Logger.Fatal(err)
	}

	groups := make(map[GroupId]*Group)
	for rows.Next() {
		var id GroupId
		var name string
		rows.Scan(&id, &name)
		groups[id] = &Group{
			Name: name,
		}
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
		Logger.Fatal(err)
	}

	//rows.Scan(&name, &description, &period, &lastCompleted)
	lastCompletedDate, _ := time.Parse("2006-01-02T15:04:05Z", lastCompleted)
	return &Task{
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
		Logger.Fatal(err)
	}
	return &Group{
		Name: name,
	}
}
