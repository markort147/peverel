package main

import (
	"fmt"
	"testing"
	"time"
)

func TestData(t *testing.T) {
	task0 := Task{
		Name:        "Vacuum living room",
		Description: "vacuuming the carpet of the living room",
		Period:      7,
	}
	task1 := Task{
		Name:        "Clean corridor floor",
		Description: "vacuuming and mopping the ground corridor",
		Period:      15,
	}
	living := Group{
		Name: "Living room",
	}
	corridor := Group{
		Name: "Corridor",
	}

	data := NewData()
	t0id := data.AddTask(&task0)
	t1id := data.AddTask(&task1)
	g0id := data.AddGroup(&living)
	g1id := data.AddGroup(&corridor)

	err := data.AddRelation(g0id, t0id)
	if err != nil {
		t.Fatal(err)
	}

	err = data.AddRelation(g1id, t1id)
	if err != nil {
		t.Fatal(err)
	}

	err = data.CompleteTask(t0id)
	if err != nil {
		t.Fatal(err)
	}

	if data.Tasks[t0id].LastCompleted.After(time.Now().Add(1*time.Second)) || data.Tasks[t0id].LastCompleted.Before(time.Now().Add(-1*time.Second)) {
		t.Fatalf("task %d, expected %v, got %v", t0id, time.Now(), data.Tasks[t0id].LastCompleted)
	}

	for g, ts := range data.Relations {
		fmt.Printf("Group: %+v\n", data.Groups[g])
		for _, t := range ts {
			fmt.Printf("\t%+v\n", data.Tasks[t])
		}
	}

}
