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

// Map functions return a slice of KeyValue.
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

// main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue, reducef func(string, []string) string) {

	// Your worker implementation here.
	// uncomment to send the Example RPC to the coordinator.
	//CallExample()
	for {
		task, nReduce, NMaps := requestTask()
		switch task.State {
		case TaskStateCompleted:
			return

		case TaskStateWait:
			time.Sleep(time.Second)

		case TaskStateIdle, TaskStateInProgress:
			if task.Tasktype == TaskMap {
				mapfunction(task, mapf, nReduce)

			} else if task.Tasktype == TaskReduce {
				reduceFunction(task, reducef, NMaps)

			}
			reportTaskComplete(task)

			/*
				// Execute the task
				if task.Tasktype == TaskMap {
					doMap(task, mapf, nReduce)
				} else if task.Tasktype == TaskReduce {
					doReduce(task, reducef, nMaps)
				}
				// Report completion
				reportTaskComplete(task)
			*/
		}

	}
}

// -------------------------------------------

func mapfunction(task Task, mapf func(string, string) []KeyValue, nReduce int) {
	// Read input file
	// Step 1: Read the input file
	content, err := os.ReadFile(task.File)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		return
		//continue TODO:
	}

	// Step 2: Call the map function - it returns key-value pairs
	// Example: mapf returns [("apple","1"), ("banana","1"), ("apple","1")]
	kva := mapf(task.File, string(content))

	// Step 3: Create empty buckets (one for each reduce worker)
	// If nReduce = 3, we create 3 empty lists
	// Partition into nReduce buckets
	buckets := make([][]KeyValue, nReduce) // ex : buckets := [[], [], []] for nReduce = 3

	// Step 4: Put each key-value pair into a bucket
	// WHY? So reduce worker 0 gets all "apple" keys, reduce worker 1 gets all "banana" keys, etc.
	for _, kv := range kva {
		// ihash(kv.Key) gives a number based on the key
		// % nReduce gives us a number between 0 and (nReduce-1)
		// This tells us which bucket (which reduce worker) gets this key
		bucket := ihash(kv.Key) % nReduce // ex : ihash("key1") % nReduce = 0
		// Add this key-value to that bucket
		buckets[bucket] = append(buckets[bucket], kv) // ex : buckets[0] = [[key1, value1], [key3, value3]]
	}
	// Now: buckets[0] = all keys that go to reduce worker 0
	//      buckets[1] = all keys that go to reduce worker 1
	//      buckets[2] = all keys that go to reduce worker 2
	// Write each bucket as JSON file "mr-X-Y"

	// Step 5: Write each bucket to a file
	for i := 0; i < nReduce; i++ {
		// Create filename: mr-{this map task}-{which reduce worker}
		fname := fmt.Sprintf("mr-%d-%d", task.Id, i)
		f, err := os.Create(fname)
		if err != nil {
			log.Fatalf("cannot create file", fname)
		}

		enc := json.NewEncoder(f)
		for _, kv := range buckets[i] {
			err := enc.Encode(&kv)
			if err != nil {
				log.Fatalf("cannot encode json")
			}
		}
		f.Close()
	}
}

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

func reportTaskComplete(task Task) {
	args := TaskCompleteArgs{
		Tasktype: task.Tasktype,
		Id:       task.Id,
	}
	reply := TaskCompleteReply{}

	call("Coordinator.TaskComplete", &args, &reply)
}

// -------------------------------------------

func requestTask() (Task, int, int) {

	args := TaskRequest{}
	reply := TaskReply{}

	ok := call("Coordinator.AssignTask", &args, &reply)
	if !ok {
		return Task{State: TaskStateCompleted}, 0, 0
	}
	return reply.Task, reply.NReduce, reply.NMaps
}

// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		//log.Fatal("dialing:", err) TODO:
		return false
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	//fmt.Println(err)
	return false
}
