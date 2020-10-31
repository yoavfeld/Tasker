package lib

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type task struct {
	Id             string
	CronSpec       string
	DockerImageURL string
	DockerCmd      []string
	DockerParams   string

	ms iManagmentStore
}

func (t *task) Run() {
	t.printf("Starting task...")
	startTime := time.Now()

	// Create docker client
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.handleError(startTime, err)
		return
	}

	// Pull task docker image
	reader, err := cli.ImagePull(ctx, t.DockerImageURL, types.ImagePullOptions{})
	if err != nil {
		t.handleError(startTime, err)
		return
	}
	io.Copy(os.Stdout, reader)

	// Create task container
	imageName := filepath.Base(t.DockerImageURL)
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Cmd:   t.DockerCmd,
	}, nil, nil, nil, "")
	if err != nil {
		t.handleError(startTime, err)
		return
	}

	// Start task contianer
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		t.handleError(startTime, err)
		return
	}

	// Wait for container to done
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		t.handleError(startTime, err)
		return
	case <-statusCh:
	}

	// Read task logs
	logReader, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		t.handleError(startTime, err)
		return
	}
	logBuffer, err := ioutil.ReadAll(logReader)
	if err != nil {
		t.handleError(startTime, err)
		return
	}
	logStr := string(logBuffer)

	// Print task logs
	t.printf(logStr)

	//Save task run to tasker management
	t.saveTaskRun(startTime, logStr, "DONE")
	t.printf("Task finished")
}

func (t *task) saveTaskRun(startTime time.Time, log string, status string) {
	taskRun := &taskRun{
		TaskId:    t.Id,
		Timestamp: startTime.String(),
		Status:    status,
		Log:       log,
	}
	t.printf("SaveTaskRun...")
	t.ms.saveTaskRun(taskRun)
}

func (t *task) handleError(startTime time.Time, err error) {
	t.printf(err.Error())
	t.saveTaskRun(startTime, err.Error(), "ERROR")
}

func (t *task) printf(msg string, params ...interface{}) {
	taskMsg := fmt.Sprintf("%s: %s", t.Id, msg)
	log.Printf(taskMsg, params...)
}
