package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type TaskState int

/*
Task states represent the lifecycle of a MapReduce task:

	TaskStateIdle: Not yet assigned to any worker
	TaskStateInProgress: Currently being executed by a worker
	TaskStateCompleted: Successfully finished
	TaskStateWait: No tasks available, worker should wait
*/
const (
	TaskStateIdle TaskState = iota // enum
	TaskStateInProgress
	TaskStateCompleted
	TaskStateWait
)

/*
Task types distinguish between map and reduce phases
*/
const (
	TaskMap    = "map"
	TaskReduce = "reduce"
)

/*
Task represents a single unit of work (map or reduce task).

	Id: Unique identifier for this task
	Tasktype: "map" or "reduce"
	File: Input file path (for map tasks only)
	State: Current execution state (idle, in-progress, completed, wait)
	StartTime: When task was assigned (for timeout detection)
*/
type Task struct {
	Id        int
	Tasktype  string
	File      string
	State     TaskState
	StartTime time.Time
}

/*
Coordinator manages the MapReduce job execution.

	mu: Mutex protecting shared state from concurrent RPC handlers
	nReduce: Number of reduce tasks, used for partitioning intermediate keys
	Files: Input files for map tasks
	mapTasks: Array of all map tasks
	reduceTasks: Array of all reduce tasks
*/
type Coordinator struct {
	mu          sync.Mutex
	nReduce     int
	Files       []string
	mapTasks    []Task
	reduceTasks []Task
}

/*
mapInnit initializes all map tasks for the coordinator.

	Creates one map task per input file with:
		- Sequential ID assignment
		- Task type set to TaskMap
		- File path for processing
		- Initial state set to TaskStateIdle
*/
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

/*
reduceInnit initializes all reduce tasks for the coordinator.

	Creates nReduce tasks with:
		Sequential ID assignment (0 to nReduce-1)
		Task type set to TaskReduce
		Initial state set to TaskStateIdle
*/
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

/*
TasksIsDone checks if all tasks in a task list are completed.

	Args: 	m (slice of tasks to check)
	Returns: true if all tasks have state TaskStateCompleted, false otherwise
*/
func (c *Coordinator) TasksIsDone(m []Task) bool {
	for _, task := range m {
		if task.State != TaskStateCompleted {
			return false
		}
	}
	return true
}

/*
AssignTask is an RPC handler that assigns tasks to workers.

	Args: 	req (empty request from worker)
			reply (with assigned task details)

	Process:
		1. Check for timed-out tasks (>10 seconds) and reset them to idle
		2. If map tasks remain, assign an idle map task
		3. Else if reduce tasks remain, assign an idle reduce task
		4. If no idle tasks available, return TaskStateWait
		5. If all work done, return TaskStateCompleted

	Mutex ensures thread-safe access to task state
*/
func (c *Coordinator) AssignTask(req *TaskRequest, reply *TaskReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for i := range c.mapTasks {
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
			task := &c.mapTasks[i]

			if task.State == TaskStateIdle {
				task.State = TaskStateInProgress
				task.StartTime = time.Now()

				reply.Task = *task // assign task to reply
				reply.NReduce = c.nReduce
				reply.NMaps = len(c.mapTasks)
				return nil
			}
		}
		reply.Task = Task{State: TaskStateWait} //dummy task to signal wait
		return nil
	} else if !c.TasksIsDone(c.reduceTasks) {
		for i := range c.reduceTasks {
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
	if c.TasksIsDone(c.mapTasks) && c.TasksIsDone(c.reduceTasks) {
		reply.Task = Task{State: TaskStateCompleted}
		fmt.Println("Coordinator: All tasks completed, signaling workers to exit")

	}

	return nil
}

/*
TaskComplete is an RPC handler for workers to report task completion.

	Args: 	args (contains task type and ID)
			reply (empty reply)

	Updates the task state to TaskStateCompleted for the specified task.
	Validates task ID is within valid range before updating.
	Mutex ensures thread-safe state modification.
*/
func (c *Coordinator) TaskComplete(args *TaskCompleteArgs, reply *TaskCompleteReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch args.Tasktype {
	case TaskMap:
		if args.Id >= 0 && args.Id < len(c.mapTasks) {
			c.mapTasks[args.Id].State = TaskStateCompleted
		}
	case TaskReduce:
		if args.Id >= 0 && args.Id < len(c.reduceTasks) {
			c.reduceTasks[args.Id].State = TaskStateCompleted
		}
	default:
		return fmt.Errorf("invalid taasktype: %s", args.Tasktype)
	}

	return nil
}

/*
server starts the RPC server thread for handling worker requests.

	Registers coordinator RPC methods and listens on domain socket.
	Socket name is generated from user ID for uniqueness.
	Runs HTTP server in background goroutine.
*/

func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()

	sockname := coordinatorSock() // get unique socket name
	os.Remove(sockname)           // if already listening on the socket, remove it first

	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

/*
Done checks if the entire MapReduce job is complete.

	Returns: true if all map and reduce tasks are completed, false otherwise
	Called periodically by main/mrcoordinator.go to determine when to exit
	Thread-safe via mutex protection
*/
func (c *Coordinator) Done() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.TasksIsDone(c.reduceTasks) && c.TasksIsDone(c.mapTasks)
}

/*
MakeCoordinator creates and initializes a new coordinator.

	Args: 	files (list of input files for map tasks)
			nReduce (number of reduce tasks to create)
	Returns: Initialized coordinator with RPC server running

	Initializes map tasks (one per file) and reduce tasks (nReduce count).
	Starts RPC server to handle worker requests.
*/
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := &Coordinator{
		nReduce: nReduce,
		Files:   files,
	}
	c.mapInnit()
	c.reduceInnit()
	c.server()
	return c
}
