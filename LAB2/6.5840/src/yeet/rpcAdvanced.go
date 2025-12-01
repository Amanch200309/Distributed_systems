package yeet

/*
TaskRequest is sent from worker to coordinator to request a new task.

	WorkerAddr: The network address of the requesting worker
*/
type TaskRequest struct {
	WorkerAddr string
}

/*
TaskReply is the coordinator's response containing task assignment.

	Task: The assigned task (or wait/completed signal)
	NReduce: Number of reduce tasks (needed for map partitioning)
	NMaps: Number of map tasks (needed for reduce file collection)
	MapWorkers: Map of map task IDs to worker addresses (for reduce workers to pull intermediate data)
*/
type TaskReply struct {
	Task       Task
	NReduce    int
	NMaps      int
	MapWorkers map[int]string
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

/*
coordinatoSock hardcoded address for the coordinator on AWS.

	Returns: ip and port of the coordinator on AWS

*/

func coordinatorSock() string {
	return "172.31.69.34" + ":8080"
}
