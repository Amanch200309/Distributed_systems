package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
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

	for {

		task, nReduce, nMaps := requestTask()
		switch task.State {
		case TaskStateCompleted:
			os.Exit(0)

		case TaskStateWait:
			time.Sleep(time.Second)

		case TaskStateInProgress:
			if task.Tasktype == TaskMap {
				mapfunction(task, mapf, nReduce)

			} else if task.Tasktype == TaskReduce {
				reduceFunction(task, reducef, nMaps)

			}
			reportTaskComplete(task)
		}
	}
}

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

		// implement atomic rename, this will handle if we crash before temp file is saved as the target file. either the new file appears fully, or only the temp file remains.
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

	ok := call(coordinatorSock(), "Coordinator.TaskComplete", &args, &reply)
	if !ok {
		fmt.Println("Worker: RPC to report task completion failed, exiting")
		os.Exit(0)
	}
}

/*
requestTask requests a new task from the coordinator.

	Returns: Task (assigned task or wait/completed signal),
			 nReduce (number of reduce tasks),
			 nMaps (number of map tasks),
			 mapWorkers (map of map task ID to worker address - ADVANCED)
	If RPC fails, returns completed state to signal worker shutdown
*/
func requestTask() (Task, int, int) {

	args := TaskRequest{}
	reply := TaskReply{}

	ok := call(coordinatorSock(), "Coordinator.AssignTask", &args, &reply)
	if !ok {
		fmt.Println("Worker: Coordinator shutdown, exiting")
		return Task{State: TaskStateCompleted}, 0, 0
	}
	return reply.Task, reply.NReduce, reply.NMaps
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
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		return false
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	return err == nil
}
