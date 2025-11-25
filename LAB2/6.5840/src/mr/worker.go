package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sort"
	"time"
)

/*
KeyValue represents a key-value pair from Map functions.

	Key: The key string (e.g., a word in word count)
	Value: The value string (e.g., "1" for each occurrence)
*/
type KeyValue struct {
	Key   string
	Value string
}

// === ADVANCED
/*
Handles RPC requests from other workers.
Used for reduce workers to fetch intermediate data from map workers.
*/
type WorkerRPC struct{}

// === END ADVANCED

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

/*
Worker is the main worker loop that requests and executes tasks from the coordinator.

	Args: 	mapf (map function: filename, content -> []KeyValue)
			reducef (reduce function: key, []values -> single output value)
	Continuously requests tasks until coordinator signals completion.
	Handles three task states:
		- TaskStateCompleted: All work done, exit
		- TaskStateWait: No tasks available, sleep and retry
		- TaskStateInProgress: Execute map or reduce task, then report completion
*/
// main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue, reducef func(string, []string) string) {

	// === ADVANCED
	workerAddr := startWorkerRPCServer()
	// === END ADVANCED

	for {
		// === ADVANCED: Pass worker address to coordinator
		task, nReduce, nMaps, mapWorkers := requestTask(workerAddr)
		// === END ADVANCED

		switch task.State {
		case TaskStateCompleted:
			return

		case TaskStateWait:
			time.Sleep(time.Second)

		case TaskStateInProgress:
			if task.Tasktype == TaskMap {
				mapfunction(task, mapf, nReduce)

			} else if task.Tasktype == TaskReduce {
				// === ADVANCED: Use pull-based reduce
				reduceFunctionAdvanced(task, reducef, nMaps, mapWorkers)
				// === END ADVANCED

			}
			reportTaskComplete(task)
		}
	}
}

// === ADVANCED: Worker RPC Server

/*
GetBucket is an RPC handler that serves intermediate data to reduce workers.

	Args: 	args (contains map task ID and reduce bucket ID)
			reply (populated with key-value pairs from the bucket)
	Returns: error (nil on success)
	Reads the intermediate file mr-{MapTaskID}-{ReduceID} and returns all key-value pairs.
*/
func (w *WorkerRPC) GetBucket(args *GetBucketArgs, reply *GetBucketReply) error {
	filename := fmt.Sprintf("mr-%d-%d", args.MapTaskID, args.ReduceID)

	file, err := os.Open(filename)
	if err != nil {
		// Empty bucket - file might not exist if bucket was empty
		reply.Data = []KeyValue{}
		return nil
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	for {
		var kv KeyValue
		if err := dec.Decode(&kv); err != nil {
			break
		}
		reply.Data = append(reply.Data, kv)
	}

	return nil
}

/*
startWorkerRPCServer starts an RPC server for this worker to serve intermediate data.

	Returns: Socket address for other workers to connect to
	Creates a Unix domain socket named "worker-{pid}.sock" and listens for RPC calls.
	Runs the HTTP server in a background goroutine.
*/
func startWorkerRPCServer() string {
	sock := fmt.Sprintf("worker-%d.sock", os.Getpid())

	rpc.Register(new(WorkerRPC))
	rpc.HandleHTTP()

	// Remove previous socket if exists
	os.Remove(sock)

	l, err := net.Listen("unix", sock)
	if err != nil {
		log.Fatalf("Worker RPC listen error: %v", err)
	}

	go http.Serve(l, nil)

	return sock
}

// === END ADVANCED

/*
mapfunction executes a map task and partitions output into intermediate files.

	Args: 	task (the map task with file to process)
			mapf (map function to apply)
			nReduce (number of reduce tasks/buckets)

	Process:
		1. Read input file content
		2. Call mapf to generate key-value pairs
		3. Partition pairs into nReduce buckets using ihash
		4. Write each bucket to intermediate file "mr-X-Y"
			where X = map task ID, Y = reduce task ID

	Example: If nReduce=3 and this is map task 5
		Creates: mr-5-0, mr-5-1, mr-5-2
		Each file contains KVs for a specific reduce worker
*/
func mapfunction(task Task, mapf func(string, string) []KeyValue, nReduce int) {
	// Step 1: Read the input file
	content, err := os.ReadFile(task.File)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		return
	}

	// Step 2: Call the map function - it returns key-value pairs
	// Example: mapf returns [("apple","1"), ("banana","1"), ("apple","1")]
	keyValuePairs := mapf(task.File, string(content))

	// Step 3: Create empty buckets (one for each reduce worker)
	// If nReduce = 3, we create 3 empty lists
	// Partition into nReduce buckets
	buckets := make([][]KeyValue, nReduce) // ex : buckets := [[], [], []] for nReduce = 3

	// Step 4: Put each key-value pair into a bucket
	// WHY? So reduce worker 0 gets all "apple" keys, reduce worker 1 gets all "banana" keys, etc.
	for _, kv := range keyValuePairs {
		bucket_index := ihash(kv.Key) % nReduce // ex : ihash("key1") % nReduce = 0
		buckets[bucket_index] = append(buckets[bucket_index], kv)
	}
	// Now: buckets[0] = all keys that go to reduce worker 0
	//      buckets[1] = all keys that go to reduce worker 1
	//      buckets[2] = all keys that go to reduce worker 2
	// Write each bucket as JSON file "mr-X-Y"

	for i := 0; i < nReduce; i++ {
		// Create filename: mr-{this map task}-{which reduce worker}
		fname := fmt.Sprintf("mr-%d-%d", task.Id, i)

		// Use atomic file creation: write to temp file, then rename
		tempFile, err := ioutil.TempFile(".", "mr-tmp-*")
		if err != nil {
			log.Fatalf("cannot create temp file for %s", fname)
		}

		enc := json.NewEncoder(tempFile)
		for _, kv := range buckets[i] {
			err := enc.Encode(&kv)
			if err != nil {
				log.Fatalf("cannot encode json")
			}
		}
		tempFile.Close()

		// Atomic rename - if we crash before this, temp file is orphaned but target file is safe
		os.Rename(tempFile.Name(), fname)
	}
}

/*
reduceFunction executes a reduce task by aggregating intermediate files.

	Args: 	task (the reduce task to execute)
			reducef (reduce function: key, []values -> output)
			nMaps (number of map tasks that produced intermediate files)

	Process:
		1. Read all intermediate files mr-*-Y where Y = this reduce task ID
		2. Collect all key-value pairs from these files
		3. Sort by key to group same keys together
		4. For each unique key, call reducef with all its values
		5. Write output to mr-out-Y using atomic rename for crash safety

	Example: For reduce task 2 with 8 map tasks
		Reads: mr-0-2, mr-1-2, mr-2-2, ..., mr-7-2
		Writes: mr-out-2

	NOTE: This is the basic implementation using shared filesystem.
*/
func reduceFunction(task Task, reducef func(string, []string) string, nMaps int) {
	// Read all intermediate files for this reduce task
	kva := []KeyValue{}
	for i := 0; i < nMaps; i++ {
		filename := fmt.Sprintf("mr-%d-%d", i, task.Id)
		file, err := os.Open(filename)
		if err != nil {
			continue // File might not exist if bucket was empty
		}

		dec := json.NewDecoder(file)
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil {
				break
			}
			kva = append(kva, kv)
		}
		file.Close()
	}

	// Sort by key
	sort.Slice(kva, func(i, j int) bool { return kva[i].Key < kva[j].Key })

	// Write output
	oname := fmt.Sprintf("mr-out-%d", task.Id)
	ofile, err := ioutil.TempFile(".", "mr-tmp-*")
	if err != nil {
		log.Fatalf("cannot create temp file")
	}
	// Call reduce function for each unique key
	i := 0
	for i < len(kva) {
		j := i + 1
		for j < len(kva) && kva[j].Key == kva[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, kva[k].Value)
		}
		output := reducef(kva[i].Key, values)

		// Write in the correct format
		fmt.Fprintf(ofile, "%v %v\n", kva[i].Key, output)
		i = j
	}
	ofile.Close()
	os.Rename(ofile.Name(), oname)
}

// === ADVANCED

/*
reduceFunctionAdvanced executes a reduce task using pull-based RPC to fetch data from map workers.

	Args: 	task (the reduce task to execute)
			reducef (reduce function: key, []values -> output)
			nMaps (number of map tasks that produced intermediate files)
			mapWorkers (map of map task ID -> worker socket address)

	Process:
		1. For each map worker, fetch intermediate data via RPC (GetBucket)
		2. Collect all key-value pairs from all map workers
		3. Sort by key to group same keys together
		4. For each unique key, call reducef with all its values
		5. Write output to mr-out-Y using atomic rename for crash safety

	Example: For reduce task 2 with 8 map tasks
		Calls: WorkerRPC.GetBucket on each of the 8 map workers
		Writes: mr-out-2
*/
func reduceFunctionAdvanced(task Task, reducef func(string, []string) string, nMaps int, mapWorkers map[int]string) {
	var kva []KeyValue

	// Fetch data from each map worker via RPC
	for m := 0; m < nMaps; m++ {
		workerAddr := mapWorkers[m]
		args := GetBucketArgs{MapTaskID: m, ReduceID: task.Id}
		reply := GetBucketReply{}

		ok := callWorker(workerAddr, "WorkerRPC.GetBucket", &args, &reply)
		if ok {
			kva = append(kva, reply.Data...)
		} else {
			log.Printf("Failed to fetch bucket from map worker %d at %s", m, workerAddr)
		}
	}

	// Sort by key
	sort.Slice(kva, func(i, j int) bool { return kva[i].Key < kva[j].Key })

	// Write output atomically
	oname := fmt.Sprintf("mr-out-%d", task.Id)
	ofile, err := ioutil.TempFile(".", "mr-tmp-*")
	if err != nil {
		log.Fatalf("cannot create temp file")
	}

	// Call reduce function for each unique key
	i := 0
	for i < len(kva) {
		j := i + 1
		for j < len(kva) && kva[j].Key == kva[i].Key {
			j++
		}

		var values []string
		for k := i; k < j; k++ {
			values = append(values, kva[k].Value)
		}

		output := reducef(kva[i].Key, values)
		fmt.Fprintf(ofile, "%v %v\n", kva[i].Key, output)

		i = j
	}

	ofile.Close()
	os.Rename(ofile.Name(), oname)
}

// === END ADVANCED

/*
reportTaskComplete notifies the coordinator that a task has finished.

	Args: 	task (completed task to report)
	Sends RPC to coordinator with task type and ID for state tracking
*/
func reportTaskComplete(task Task) {
	args := TaskCompleteArgs{
		Tasktype: task.Tasktype,
		Id:       task.Id,
	}
	reply := TaskCompleteReply{}

	call(coordinatorSock(), "Coordinator.TaskComplete", &args, &reply)
}

/*
requestTask requests a new task from the coordinator.

	Returns: Task (assigned task or wait/completed signal),
			 nReduce (number of reduce tasks),
			 nMaps (number of map tasks),
			 mapWorkers (map of map task ID to worker address - ADVANCED)
	If RPC fails, returns completed state to signal worker shutdown
*/
func requestTask(workerAddr string) (Task, int, int, map[int]string) {

	args := TaskRequest{WorkerAddr: workerAddr}
	reply := TaskReply{}

	ok := call(coordinatorSock(), "Coordinator.AssignTask", &args, &reply)
	if !ok {
		return Task{State: TaskStateCompleted}, 0, 0, nil
	}
	return reply.Task, reply.NReduce, reply.NMaps, reply.MapWorkers
}

/*
call sends an RPC request to the coordinator and waits for response.

	Args: 	sockname (Unix domain socket path for coordinator connection)
			rpcname (RPC method name, e.g., "Coordinator.AssignTask")
			args (request arguments)
			reply (response struct to populate)
	Returns: true if RPC succeeds, false if connection/call fails
*/
func call(sockname string, rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		return false
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	return false
}

// === ADVANCED: Worker-to-Worker RPC

/*
callWorker sends an RPC request to another worker and waits for response.

	Args: 	sockname (Unix domain socket path for worker connection)
			rpcname (RPC method name, e.g., "WorkerRPC.GetBucket")
			args (request arguments)
			reply (
			response struct to populate)
	Returns: true if RPC succeeds, false if connection/call fails
	Used by reduce workers to fetch intermediate data from map workers
*/
func callWorker(sockname string, rpcname string, args interface{}, reply interface{}) bool {
	return call(sockname, rpcname, args, reply)
}

// === END ADVANCED
