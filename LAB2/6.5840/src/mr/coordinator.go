package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type TaskState int

// Task states
const (
	TaskStateIdle TaskState = iota
	TaskStateInProgress
	TaskStateCompleted
	TaskStateWait
)

const (
	TaskMap    = "map"
	TaskReduce = "reduce"
)

type Task struct {
	Id        int
	Tasktype  string // map or reduce
	File      string
	State     TaskState // idle, in-progress, completed, wait
	StartTime time.Time // time when the task was assigned
}

type Coordinator struct {
	// Your definitions here.
	mu          sync.Mutex // mutex for synchronizing access to shared data
	nReduce     int        // number of reduce tasks
	Files       []string
	mapTasks    []Task
	reduceTasks []Task
}

// Your code here -- RPC handlers for the worker to call.

// initialize map tasks for the coordinator
func (c *Coordinator) mapInnit() {

	for i, f := range c.Files {
		task := Task{
			Id:       i,
			Tasktype: TaskMap,
			File:     f,
			State:    TaskStateIdle,
		}
		c.mapTasks = append(c.mapTasks, task)
	}
}

// initialize reduce tasks for the coordinator
func (c *Coordinator) reduceInnit() {
	for i := 0; i < c.nReduce; i++ {
		task := Task{
			Id:       i,
			Tasktype: TaskReduce,
			State:    TaskStateIdle,
		}
		c.reduceTasks = append(c.reduceTasks, task)
	}
}

func (c *Coordinator) TasksIsDone(m []Task) bool {
	//return len(m) == 0 TODO: kanske ändra sen
	for _, task := range m {
		if task.State != TaskStateCompleted {
			return false
		}
	}
	return true
}

func (c *Coordinator) assignTask(req *TaskRequest, reply *TaskReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for i := range c.mapTasks { // TODO: give task to different worker? See if works
		if c.mapTasks[i].State == TaskStateInProgress && now.Sub(c.mapTasks[i].StartTime) > 10*time.Second {
			c.mapTasks[i].State = TaskStateIdle
		}
	}
	for i := range c.reduceTasks {
		if c.reduceTasks[i].State == TaskStateInProgress && now.Sub(c.reduceTasks[i].StartTime) > 10*time.Second {
			c.reduceTasks[i].State = TaskStateIdle
		}
	}

	if !c.TasksIsDone(c.mapTasks) {

		for i := range c.mapTasks {
			// hämta task
			task := &c.mapTasks[i]

			if task.State == TaskStateIdle {
				task.State = TaskStateInProgress
				task.StartTime = time.Now()
				reply.Task = *task            // assign task to reply
				reply.NReduce = c.nReduce     // ADD - map needs to create NReduce intermediate files
				reply.NMaps = len(c.mapTasks) // ADD - for completeness
				return nil
			}
		}
		reply.Task = Task{State: TaskStateWait}
		return nil

	} else if !c.TasksIsDone(c.reduceTasks) {
		for i := range c.reduceTasks {
			// hämta task
			task := &c.reduceTasks[i]

			if task.State == TaskStateIdle {
				task.State = TaskStateInProgress
				task.StartTime = time.Now()
				reply.Task = *task // assign task to reply
				reply.NReduce = c.nReduce
				reply.NMaps = len(c.mapTasks)
				return nil
			}
		}
		reply.Task = Task{State: TaskStateWait}
		return nil
	}
	if c.Done() {
		reply.Task = Task{State: TaskStateCompleted}
	}

	return nil
}

// RPC handler for workers to report task completion
func (c *Coordinator) TaskComplete(args *TaskCompleteArgs, reply *TaskCompleteReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if args.Tasktype == TaskMap {
		if args.Id >= 0 && args.Id < len(c.mapTasks) {

			c.mapTasks[args.Id].State = TaskStateCompleted
		}
	} else if args.Tasktype == TaskReduce {
		if args.Id >= 0 && args.Id < len(c.reduceTasks) {

			c.reduceTasks[args.Id].State = TaskStateCompleted
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
func (c *Coordinator) Done() bool { // TODO: Implementation of task is done will be noticed here!
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.TasksIsDone(c.reduceTasks) && c.TasksIsDone(c.mapTasks)
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := &Coordinator{
		nReduce: nReduce,
		Files:   files,
	}
	c.mapInnit()
	c.reduceInnit()
	// Your code here.

	c.server()
	return c
}
