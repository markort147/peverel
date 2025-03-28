package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/markort147/gopkg/log"
	"time"
)

type PsqlData struct {
	*sql.DB
}

func (p *PsqlData) Init(cfg *Config) {
	log.Logger.Debugf("opening database connection %q", cfg.Database.ConnStr)
	db, err := sql.Open("postgres", cfg.Database.ConnStr)
	//db, err := sql.Open("postgres", "postgresql://peverel:peverel@localhost:5433/peverel?sslmode=disable")
	if err != nil {
		log.Logger.Fatal(err)
	}
	p.DB = db
}

func (p *PsqlData) AddTask(task *Task) (id TaskId) {
	err := p.DB.QueryRow("INSERT into tasks (name, description, period, last_completed) VALUES ($1, $2, $3, $4) RETURNING id",
		task.Name,
		task.Description,
		task.Period,
		task.LastCompleted,
	).Scan(&id)
	if err != nil {
		log.Logger.Fatal(err)
	}
	//defer rows.Close()
	//if err != nil {
	//	log.Logger.Fatal(err)
	//}

	//last, _ := rows.LastInsertId()
	//id = TaskId(last)
	//rows.Scan(&id)
	return id
}

func (p *PsqlData) AddGroup(group *Group) (id GroupId) {
	err := p.DB.QueryRow("INSERT into groups (name) VALUES ($1) RETURNING id",
		group.Name,
	).Scan(&id)
	//defer rows.Close()
	if err != nil {
		log.Logger.Fatal(err)
	}

	//last, _ := rows.LastInsertId()
	//id = GroupId(last)
	//rows.Scan(&id)
	return id
}

func (p *PsqlData) CompleteTask(id TaskId) error {
	_, err := p.DB.Exec("UPDATE tasks SET last_completed=$1 WHERE id=$2", time.Now(), id)
	return err
}

func (p *PsqlData) AddRelation(groupId GroupId, taskIds ...TaskId) error {
	for _, taskId := range taskIds {
		log.Logger.Debugf("adding task %d to group %d", taskId, groupId)
		_, err := p.DB.Exec("UPDATE tasks SET group_id=$1 WHERE id=$2", groupId, taskId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PsqlData) GetTasksByGroup(groupId GroupId) map[TaskId]*Task {
	rows, err := p.DB.Query("SELECT id, name, description, period, last_completed FROM tasks where group_id=$1", groupId)
	defer rows.Close()
	if err != nil {
		log.Logger.Fatal(err)
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

func (p *PsqlData) GetUnassignedTasks() map[TaskId]*Task {
	rows, err := p.DB.Query("SELECT id, name, description, period, last_completed FROM tasks where group_id is NULL")
	defer rows.Close()
	if err != nil {
		log.Logger.Fatal(err)
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

func (p *PsqlData) GetTasks() map[TaskId]*Task {
	rows, err := p.DB.Query("SELECT id, name, description, period, last_completed FROM tasks")
	defer rows.Close()
	if err != nil {
		log.Logger.Fatal(err)
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
		log.Logger.Debugf("LastCompleted: %s - LastCompletedDate: %v", lastCompleted, lastCompletedDate)
		tasks[id] = &Task{
			Name:          name,
			Description:   description,
			Period:        period,
			LastCompleted: lastCompletedDate,
		}
		log.Logger.Debugf("Found task %+v", tasks[id])
	}

	return tasks
}

func (p *PsqlData) GetGroups() map[GroupId]*Group {
	rows, err := p.DB.Query("SELECT id, name from groups")
	defer rows.Close()
	if err != nil {
		log.Logger.Fatal(err)
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

func (p *PsqlData) GetTask(id TaskId) *Task {
	var name string
	var description string
	var period int
	var lastCompleted string
	err := p.DB.QueryRow("SELECT name, description, period, last_completed FROM tasks WHERE id=$1", id).Scan(&name, &description, &period, &lastCompleted)
	//defer rows.Close()
	if err != nil {
		log.Logger.Fatal(err)
	}

	//rows.Scan(&name, &description, &period, &lastCompleted)
	lastCompletedDate, _ := time.Parse("2006-01-02T15:04:05Z", lastCompleted)
	log.Logger.Debugf("LastCompleted: %s - LastCompletedDate: %v", lastCompleted, lastCompletedDate)
	return &Task{
		Name:          name,
		Description:   description,
		Period:        period,
		LastCompleted: lastCompletedDate,
	}
}
