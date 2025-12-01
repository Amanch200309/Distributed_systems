package mr

import (
	"os"
	"strconv"
)

/*
TaskRequest is sent from worker to coordinator to request a new task.

	Currently empty - no arguments needed for task requests
*/
type TaskRequest struct {
}

/*
TaskReply is the coordinator's response containing task assignment.

	Task: The assigned task (or wait/completed signal)
	NReduce: Number of reduce tasks (needed for map partitioning)
	NMaps: Number of map tasks (needed for reduce file collection)
*/
type TaskReply struct {
	Task    Task
	NReduce int
	NMaps   int
}

/*
TaskCompleteArgs is sent from worker to coordinator to report task completion.

	Tasktype: "map" or "reduce"
	Id: Task identifier (index in mapTasks or reduceTasks array)
*/
type TaskCompleteArgs struct {
	Tasktype string
	Id       int
}

/*
TaskCompleteReply is the coordinator's acknowledgment of task completion.

	Currently empty - no return data needed
*/
type TaskCompleteReply struct {
}

/*
coordinatorSock generates a unique Unix domain socket name for RPC communication.

	Returns: Socket path in /var/tmp with user ID suffix
	Format: /var/tmp/5840-mr-<uid>
*/

func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
