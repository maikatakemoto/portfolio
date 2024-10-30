package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

// example to show how to declare the arguments
// and reply for an RPC.
type ExampleArgs struct {
	X int
}
type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.

type TaskArgs struct {
}

type TaskReply struct {
	TaskType   string
	TaskStatus int

	MapFile    string   // Which map file we are reading from --> 1 at a time 
	
	MapID 	   int 		// Keep track of which map task we're on 
	ReduceID   int 		// Keep track of which reduce task we're on 
	Files      []string // The files we output (i.e. intermediate / final ofile)

	NReduce int
}

type MapSuccessArgs struct {
	File string
}

type SuccessReply struct {
}

type ReduceSuccessArgs struct {
	NReduceind int
}

type IntermediateFile struct {
	Filename string
	NReduceind int
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
