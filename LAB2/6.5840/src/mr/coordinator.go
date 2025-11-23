package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"
)

type TaskState int

// Task states
const (
	TaskStateIdle TaskState = iota
	TaskStateInProgress
	TaskStateCompleted
)

const (
	TaskMap    = "map"
	TaskReduce = "reduce"
)

type Task struct {
	Tasktype  string // map or reduce
	File      string
	State     TaskState // idle, in-progress, completed
	StartTime time.Time // time when the task was assigned
}

type Coordinator struct {
	// Your definitions here.
	nReduce     int // number of reduce tasks
	Files       []string
	mapTasks    []Task
	reduceTasks []Task
}

// Your code here -- RPC handlers for the worker to call.

func (c *Coordinator) TasksIsDone(m []Task) bool {
	return len(m) == 0
}

func (c *Coordinator) ExampleTwo(req *TaskRequest, reply *TaskReply) error {

	if !c.TasksIsDone(c.mapTasks) {

		for i := range c.mapTasks {
			// hämta task
			task := &c.mapTasks[i]

			if task.State == TaskStateIdle {
				task.State = TaskStateInProgress
				reply.Task = *task // assign task to reply
			}
		}
	}
	if !c.TasksIsDone(c.reduceTasks) {
		for i := range c.reduceTasks {
			// hämta task
			task := &c.reduceTasks[i]

			if task.State == TaskStateIdle {
				task.State = TaskStateInProgress
				reply.Task = *task // assign task to reply
			}
		}
	}
	return nil
}

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	ret := false

	// Your code here.

	return ret
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// Your code here.

	c.server()
	return &c
}
