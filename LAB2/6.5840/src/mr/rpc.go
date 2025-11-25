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
	WorkerAddr string
}

/*
TaskReply is the coordinator's response containing task assignment.

	Task: The assigned task (or wait/completed signal)
	NReduce: Number of reduce tasks (needed for map partitioning)
	NMaps: Number of map tasks (needed for reduce file collection)
*/
type TaskReply struct {
	Task       Task
	NReduce    int
	NMaps      int
	MapWorkers map[int]string
}

/*
DoneRequest is sent from worker to check if job is complete.

	Currently unused in implementation
*/
type DoneRequest struct {
}

/*
DoneReply is the coordinator's response about job completion.

	Currently unused in implementation
*/
type DoneReply struct {
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

// ==================== ADVANCED: Worker-to-Worker RPC ====================

/*
GetBucketArgs is sent from reduce worker to map worker to request intermediate data.

	MapTaskID: ID of the map task that produced the data
	ReduceID: ID of the reduce task requesting the data (bucket number)
*/
type GetBucketArgs struct {
	MapTaskID int
	ReduceID  int
}

/*
GetBucketReply contains intermediate key-value pairs for a specific reduce bucket.

	Data: Array of key-value pairs from mr-{MapTaskID}-{ReduceID} file
*/
type GetBucketReply struct {
	Data []KeyValue
}

// ==================== END ADVANCED ====================

/*
coordinatorSock generates a unique Unix domain socket name for RPC communication.

	Returns: Socket path in /var/tmp with user ID suffix
	Uses /var/tmp instead of current directory for AFS compatibility
	Format: /var/tmp/5840-mr-<uid>
*/
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
