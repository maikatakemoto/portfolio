package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	//"io"
	"log"
	"net/rpc"
	"os"
	"sort"
)

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

// i.e. (Key, Value) pair
type KeyValue struct {
	Key   string
	Value string
}

// use ihash(key) mod(%) NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// mr-main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	for {
		reply := CallCoordinator()

		switch {
		case reply.TaskType == "Map":
			// Execute Map task
			file, err := os.ReadFile(reply.MapFile)

			if err != nil {
				log.Printf("Error reading file.")
			}

			kva := mapf(reply.MapFile, string(file)) // Output of 1 mapTask
			// println(len(kva))

			for i := 0; i < reply.NReduce; i++ { // Divide mapTask into nReducers
				name := fmt.Sprintf("mr-%v-%v", reply.MapID, i)
				// println(name)
				output, err := os.Create(name) // Create empty file

				if err != nil {
					log.Printf("Error creating intermediate files.")
				}

				//defer output.Close()
				enc := json.NewEncoder(output)

				// Hash and encode intermediate files
				for _, kvintermediate := range kva {
					partitionID := ihash(kvintermediate.Key) % reply.NReduce // Creating partitionID
					// Check if (key, value) pair belongs in current partition
					if partitionID == i {
						err := enc.Encode(&kvintermediate)

						if err != nil {
							break
						}
					}
				}
				output.Close()

			}
			MapTaskDone(reply.MapFile)
			// Send an RPC call back if Map task is successfully completed

		case reply.TaskType == "Reduce":
			intermediate := []KeyValue{} // Going to be used to store all (key, value) pairs read from a file
			// println(len(reply.Files))
			for h := 0; h < len(reply.Files); h+=1 {
				filename := fmt.Sprintf("mr-%v-%d", h, reply.ReduceID)
				//log.Printf("filenane %v", filename)
				file, err:= os.Open(filename)

				if err != nil {
					log.Printf("Cannot open %v", filename)
				}

				dec := json.NewDecoder(file)
				for {
					var kv KeyValue
					err := dec.Decode(&kv)

					// No more (key, value) pairs to read
				if err != nil {
						break
					}
					intermediate = append(intermediate, kv)
				}
				file.Close()	
			}
			
			sort.Sort(ByKey(intermediate))

			oname := fmt.Sprintf("mr-out-%v", reply.ReduceID)
			ofile, _ := os.Create(oname)

			// Do the Reduce task for each key
			i := 0
			for i < len(intermediate) {
				j := i + 1
				for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
					j++
				}
				values := []string{}
				for k := i; k < j; k++ {
					values = append(values, intermediate[k].Value)
				}
				output := reducef(intermediate[i].Key, values)
				fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output)
				i = j
			}
			ofile.Close()
			ReduceTaskDone(reply.ReduceID)
		}
	}
}

func CallCoordinator() TaskReply {
	// Request the tasks
	args := TaskArgs{}
	reply := TaskReply{}
	call("Coordinator.DistTasks", &args, &reply)
	return reply
}

func MapTaskDone(filename string) {
	// Notify the Coordinator if MapTask is succesfully completed
	// println("Notifying the coordinator that a map task is done!")
	args := MapSuccessArgs{filename}
	reply := TaskReply{}
	call("Coordinator.MapTaskSuccess", &args, &reply)
}

func ReduceTaskDone(reduceTask int) {
	// Notify the Coordinator if a reduce task is succesfully completed
	args := ReduceSuccessArgs{reduceTask}
	reply := TaskReply{}
	call("Coordinator.ReduceTaskSuccess", &args, &reply)
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
// DO NOT MODIFY
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println("Unable to Call", rpcname, "- Got error:", err)
	return false
}
