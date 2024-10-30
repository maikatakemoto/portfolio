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

const (
	InProgress = iota
	NotStarted
	Done
)

type Coordinator struct {
	initial_files         []string         // Initial files it hands out to workers
	mapTaskTracker    	 map[string]int   // Map to keep track of which Map tasks are in progress
	reduceTaskTracker 	 map[int]int      // Map to keep track of which Reduce tasks are in progress
	nReduce              int
	finished 		     bool
	lock                 *sync.Mutex
	mapCounter			 int      		  // Keep track of how many map tasks have been executed 
	reduceCounter        int 			  // Keep track of how many reduce tasks have been executed
}

type MapTask struct {
	MapID int
	file  string
}

// Initialize channels for Map and Reduce tasks
var mapTasks chan MapTask
var reduceTasks chan int

// Your code here -- RPC handlers for the worker to call.
func (c *Coordinator) DistTasks(args *TaskArgs, reply *TaskReply) error {
	// Distribute Map and Reduce tasks to the Workers
	c.lock.Lock()
	defer c.lock.Unlock()
	
	select {
	case MapTask := <-mapTasks:
		go c.FaultToleranceMap(MapTask) // Launch Go routine to make sure Worker gets task done in 10 seconds or less
		
		reply.MapID = MapTask.MapID
		reply.TaskType = "Map"
		reply.MapFile = MapTask.file
		reply.NReduce = c.nReduce
		reply.Files = c.initial_files

		c.mapTaskTracker[MapTask.file] = InProgress

		return nil

	case reduceTask := <-reduceTasks:
		go c.FaultToleranceReduce(reduceTask)
		
		reply.TaskType = "Reduce"
		reply.ReduceID = reduceTask
		reply.NReduce = c.nReduce
		reply.Files = c.initial_files 

		c.reduceTaskTracker[reduceTask] = InProgress
		
	default:
		if c.mapCounter == len(c.initial_files) && c.reduceCounter == c.nReduce {
			c.finished = true
		}
	}
	return nil
}

// mr-main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {

	ret := c.finished 

	return ret
}

// create a Coordinator.
// mr-main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{
		mapCounter: 0,
		reduceCounter: 0,
		initial_files: files,
		nReduce: nReduce,
	}

	// Create channels to collect Map and Reduce tasks
	mapTasks = make(chan MapTask, len(files))
	reduceTasks = make(chan int, nReduce)

	// Create maps to keep track of task successes
	c.mapTaskTracker = make(map[string]int, len(files))
	c.reduceTaskTracker = make(map[int]int, nReduce)

	// Initialize every Task to be set false since not complemeted yet
	for index, file := range files {
		c.mapTaskTracker[file] = NotStarted
		mapTask := MapTask{}
		mapTask.MapID = index
		mapTask.file = file
		mapTasks <- mapTask // Send Map tasks to channel
	}

	c.lock = &sync.Mutex{}
	c.server()
	return &c
}

func (c *Coordinator) FaultToleranceMap(task MapTask) {
	// Watch the Workers on a 10 sec ticker
	// If timed out, reassign task to another worker
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	select {
	case <-ticker.C:
		c.lock.Lock()
		status := c.mapTaskTracker[task.file]
		c.lock.Unlock()

		// Task is InProgress after 10 sec --> Worker failure --> Reassign task
		// Task status needs to be reset
		if status != Done {
			c.lock.Lock()
			c.mapCounter = c.mapCounter - 1
			c.mapTaskTracker[task.file] = NotStarted
			c.lock.Unlock()
			
			mapTasks <- task // Send task back to channel
		} 

	default:
		return
	}
}

func (c *Coordinator) MapTaskSuccess(args *MapSuccessArgs, reply *SuccessReply) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.mapCounter++
	c.mapTaskTracker[args.File] = Done
	// Check to see if all Map tasks are done 
	// If so, start Reduce by initializing every task to NotStarted
	if c.mapCounter == len(c.initial_files) {
		for i := 0; i < c.nReduce; i++ {
			c.reduceTaskTracker[i] = NotStarted
			reduceTasks <- i // Send Reduce tasks to Reduce channel
		}
	}
	return nil
}

func (c *Coordinator) FaultToleranceReduce(reduceTask int) {
	// Watch the Workers on a 10 sec ticker
	// If timed out, reassign task to another worker
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	select {
	case <-ticker.C:
		c.lock.Lock()
		status := c.reduceTaskTracker[reduceTask]
		c.lock.Unlock()

		if status != Done {
			c.lock.Lock()
			c.reduceCounter = c.reduceCounter - 1
			c.reduceTaskTracker[reduceTask] = NotStarted
			c.lock.Unlock()
			
			reduceTasks <- reduceTask
		} 

	default:
		return
	}
}

func (c *Coordinator) ReduceTaskSuccess(args *ReduceSuccessArgs, reply *SuccessReply) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.reduceCounter++ // Increment the number of Reduce tasks that are done
	
	if c.reduceCounter == c.nReduce {
		c.finished = true
		return nil
	}
	return nil
}

// start a thread that listens for RPCs from worker.go
// DO NOT MODIFY
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
