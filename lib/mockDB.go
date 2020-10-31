package lib

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type mockDB struct {
}

func NewMockDB() *mockDB {
	return &mockDB{}
}

func (m *mockDB) getTasks() ([]*task, error) {
	file, err := ioutil.ReadFile("tasksTableMock.json")
	if err != nil {
		return nil, err
	}
	var tasks []*task
	err = json.Unmarshal([]byte(file), &tasks)
	return tasks, err
}

func (m *mockDB) saveTaskRun(tr *taskRun) {
	log.Printf("Saving task run: %v (mock only)", tr)
}
