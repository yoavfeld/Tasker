package lib

import (
	"fmt"
	"log"
	"net/http"

	"github.com/robfig/cron"
)

type Server struct {
	conf  *Config
	ms    iManagmentStore
	tasks map[string]*task
	cron  *cron.Cron
}

type iManagmentStore interface {
	getTasks() ([]*task, error)
	saveTaskRun(tr *taskRun)
}

func NewServer(conf *Config) *Server {
	return &Server{
		conf: conf,
		ms:   NewMockDB(),
	}
}

func (s *Server) Start() error {

	// Get tasks from managment DB
	cr := cron.New()
	tasks, err := s.ms.getTasks()
	if err != nil {
		return err
	}

	// Start tasks as cron jobs
	s.tasks = make(map[string]*task)
	for _, task := range tasks {
		log.Printf("Init task: %+v", task)
		task.ms = s.ms
		_, err := cr.AddFunc(task.CronSpec, task.Run)
		if err != nil {
			return err
		}
		s.tasks[task.Id] = task
	}
	s.cron = cr
	cr.Start()
	log.Print("Tasker server stared successfully")

	// Start Tasker http listener
	s.httpListener()
	return nil
}

// http listener to execute tasks on demand
func (s *Server) httpListener() {
	http.HandleFunc("/runTask", s.runTask)
	http.ListenAndServe(":"+s.conf.Port, nil)
}

func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/runTask":
		s.runTask(res, req)
	default:
		res.WriteHeader(http.StatusNotFound)
	}
}

func (s *Server) runTask(res http.ResponseWriter, req *http.Request) {
	taskId := req.FormValue("id")
	task, ok := s.tasks[taskId]
	if !ok {
		errMsg := fmt.Sprintf("task id %s was not found", taskId)
		log.Printf(errMsg)
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(errMsg))
		return
	}
	log.Printf("Running task id %s by http trigger", taskId)
	go task.Run()
}
