package raft

//
// This is an outline of the API that raft must expose to
// the service (or tester). See comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   Create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   Start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   Each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester) in the same server.
//

import (
	"bytes"
	// "fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"cs350/labgob"
	"cs350/labrpc"
)

const (
	Follower = iota
	Candidate
	Leader
)

// Need two timeouts:
// 1. Election timeout: The amount of time a Follower waits until starting a new election.
// After election timeout, Follower becomes a Candidate, starts a new term, votes for itself,
// and sends out RequestVote RPCs. The reciever will reset their election timeout once they vote.
// 2. Heartbeat timeout: How often the Leader sends out heartbeats

const (
	MinTimeout = 500
	MaxTimeout = 800
)

// Handling election timeout
func randTimeout() time.Duration {
	TimeOut := MinTimeout + rand.Intn(MaxTimeout-MinTimeout)
	return time.Duration(TimeOut) * time.Millisecond
}

// As each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). Set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // This peer's index into peers[]
	dead      int32               // Set by Kill()
	state     int                 // If the server is a Follower, Candidate, or Leader

	// Your data here (4A, 4B).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

	applyMsg chan ApplyMsg

	// Persistent state on all servers
	currentTerm int        // Latest term server has seen (initalize to 0)
	votedFor    int        // CandidateID that recieved vote in current term
	log         []LogEntry // Array to hold all log entries

	//Volatile state on all servers
	commitIndex int // Index of highest log entry known to be committed

	// Volatile state on leaders (reinitialized after election)
	nextIndex  []int // Index of the next log entry to send to that server
	matchIndex []int // Index of highest log entry known to be replicated on server

	lastHeartbeat  time.Time // Last time a Leader sent out a heartbeat
	lastEntryIndex int       // Index of Candidate's last log entry
}

func getMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	var term int
	var isleader bool

	// Your code here (4A).
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if rf.state == Leader {
		isleader = true
	} else {
		isleader = false
	}

	term = rf.currentTerm

	return term, isleader
}

// Save Raft's persistent state to stable storage, where it
// can later be retrieved after a crash and restart. See paper's
// Figure 2 for a description of what should be persistent.
func (rf *Raft) persist() {
	// Your code here (4B).
	// Example:
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	// fmt.Println("Server", rf.me, "encoding Log looks like this:", rf.log)
	e.Encode(rf.log)
	e.Encode(rf.currentTerm)
	e.Encode(rf.votedFor)
	data := w.Bytes()
	rf.persister.SaveRaftState(data)
}

// Restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (4B).
	// Example:
	rf.mu.Lock()
	defer rf.mu.Unlock()
	r := bytes.NewBuffer(data)
	d := labgob.NewDecoder(r)
	currentTerm := 0
	votedFor := 0
	if d.Decode(&rf.log) != nil ||
		d.Decode(&currentTerm) != nil || d.Decode(&votedFor) != nil {
		// fmt.Println("Error decoding data")
	} else {
		rf.lastEntryIndex = len(rf.log) - 1
		rf.currentTerm = currentTerm
		rf.votedFor = votedFor
		
		// fmt.Println("I'm server", rf.me, "and my log looks like this:", rf.log, "with last entry index:", rf.lastEntryIndex, "and votedFor", rf.votedFor)
	}
}

// Example RequestVote RPC arguments structure.
// Field names must start with capital letters!
type RequestVoteArgs struct {
	// Your data here (4A, 4B).
	Term         int // Candidate's term
	CandidateID  int // Candidate requesting vote
	LastLogIndex int // Index of candidate's last log entry
	LastLogTerm  int // Term of candidate's last log entry
}

// Example RequestVote RPC reply structure.
// Field names must start with capital letters!
type RequestVoteReply struct {
	// Your data here (4A).
	Term        int  // currentTerm for candidate to update itself
	VoteGranted bool // True means candidate recieved vote
}

type LogEntry struct {
	Term    int // The Leader's term this log entry corresponds to
	Command interface{}
}

// Example RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (4A, 4B).
	// Issue: Server voting for Candidate's whose logs are not at least as up-to-date
	rf.mu.Lock()
	defer rf.mu.Unlock()
	defer rf.persist()

	reply.VoteGranted = false
	reply.Term = rf.currentTerm


	// fmt.Println("Candidate's term:", args.Term, "Server", rf.me, "term:", rf.currentTerm)

	if args.Term < rf.currentTerm {
		return
	}

	if args.Term > rf.currentTerm {
		// fmt.Println("Server", rf.me, "is now a follower because of outdated term from RequestVote")
		rf.state = Follower
		rf.votedFor = -1
		rf.currentTerm = args.Term
	}

	// Check if Candidate's log is at least as up-to-date as reciever's
	Ahead := args.LastLogTerm > rf.log[rf.lastEntryIndex].Term
	SameTerm := args.LastLogTerm == rf.log[rf.lastEntryIndex].Term
	SameIndex := args.LastLogIndex >= rf.lastEntryIndex
	UpToDate := Ahead || (SameIndex && SameTerm)

	if (rf.votedFor == -1 || rf.votedFor == args.CandidateID) && UpToDate {
		reply.VoteGranted = true
		rf.votedFor = args.CandidateID
		rf.lastHeartbeat = time.Now()
		// fmt.Println("Server", rf.me, "voted for server", args.CandidateID, "because log was up to date")
	} else {
		reply.VoteGranted = false
		// fmt.Println("Server", rf.me, "did not vote for server", args.CandidateID, "because", rf.me, "has already voted for", rf.votedFor, "or log is not up to date")
	}
}

// Example code to send a RequestVote RPC to a server.
// Server is the index of the target server in rf.peers[].
// Expects RPC arguments in args. Fills in *reply with RPC reply,
// so caller should pass &reply.
//
// The types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// Look at the comments in ../labrpc/labrpc.go for more details.
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

type AppendEntriesArgs struct {
	Term         int        // Leader's term
	LeaderID     int        // So Follower can redirect clients
	PrevLogIndex int        // Index of log entry immediately preceding the new ones
	PrevLogTerm  int        // Term of PrevLogIndex entry
	Entries      []LogEntry // Log entries to store (empty for heartbeat)
	LeaderCommit int        // Leader's commit index
}

type AppendEntriesReply struct {
	Term          int  // currentTerm, for Leader to update itself
	Success       bool // Return true if Follower contained entry matching PrevLogIndex and PrevLogTerm
	ConflictTerm  int  // Term of conflicting entry, if any
	ConflictIndex int  // Index of conflicting entry, if any
}

// Resets the election timeout so that other servers don't step forward as leaders
// when one has already been elected.
func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	reply.ConflictIndex = -1
	reply.ConflictTerm = -1
	reply.Success = false

	if args.Term < rf.currentTerm || len(rf.log) < args.PrevLogIndex + 1 {
		return
	}

	if args.LeaderCommit < rf.commitIndex {
		return
	}

	// Log doesn't contain entry at PrevLogIndex whose term matches PrevLogTerm
	if rf.log[args.PrevLogIndex].Term != args.PrevLogTerm {
		// fmt.Println("Server", rf.me, "conflict detected at term", args.PrevLogTerm)
		// Rollback
		reply.ConflictTerm = rf.log[args.PrevLogIndex].Term
		for j := args.PrevLogIndex; j >= 0; j-- {
			if rf.log[j].Term != reply.ConflictTerm {
				reply.ConflictIndex = j
				break
			}
		}
		rf.lastHeartbeat = time.Now()
		return
	}

	// Check if entries conflict (same index but different terms)
	i := 0
	for ; i < len(args.Entries); i++ {
		// fmt.Println("Server", rf.me, "checking entry at index", args.PrevLogIndex+1+i)
		current := args.PrevLogIndex + 1 + i
		if current > len(rf.log) - 1 { // Log isn't long enough to append new entries
			break
		}
		if rf.log[current].Term != args.Entries[i].Term {
			// fmt.Println("Server", rf.me, "conflict detected at index", current, "and length of log is", len(rf.log))
			// Delete existing entry and all that follow it
			rf.log = rf.log[:current]
			rf.lastEntryIndex = len(rf.log) - 1
			// fmt.Println("The length of the log after deleting conflicting entries is:", len(rf.log), "with lastEntryIndex:", rf.lastEntryIndex)
			rf.lastHeartbeat = time.Now()
			rf.persist()
			return
		}
	}
	reply.Success = true
	rf.lastHeartbeat = time.Now()

	// Append new entries
	if len(args.Entries) > 0 {
		rf.log = append(rf.log, args.Entries[i:]...)
		// fmt.Println("Server", rf.me, "appended entries to log", rf.log)
		rf.lastEntryIndex = len(rf.log) - 1
		// fmt.Println("Now server's last entry index is:", rf.lastEntryIndex)
		reply.Success = true
		rf.lastHeartbeat = time.Now()
		rf.persist()
	}

	// Committing new entries
	// If leader commits are ahead of server commits, server must commit everything
	// it hasn't committed to catch up to the leader
	// fmt.Println("Leader's commit index:", args.LeaderCommit, "Server", rf.me, "'s commit index:", rf.commitIndex)
	if args.LeaderCommit > rf.commitIndex {
		min := getMin(args.LeaderCommit, rf.lastEntryIndex)
		for i := rf.commitIndex + 1; i <= min; i++ {
			rf.commitIndex = i
			rf.applyMsg <- ApplyMsg{
				CommandValid: true,
				Command:      rf.log[i].Command,
				CommandIndex: i,
			}
		}
		rf.lastHeartbeat = time.Now()
		rf.persist()
		// fmt.Println("Server", rf.me, "committed entries up to index", rf.commitIndex)
		return
	} 
}

// Send out heartbeats periodically to the servers
func (rf *Raft) SendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

// The service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. If this
// server isn't the leader, returns false. Otherwise start the
// agreement and return immediately. There is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. Even if the Raft instance has been killed,
// this function should return gracefully.
//
// The first return value is the index that the command will appear at
// if it's ever committed. The second return value is the current
// term. The third return value is true if this server believes it is
// the leader.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (4B).
	// fmt.Println("Server", rf.me, "received a command", command)
	rf.mu.Lock()
	defer rf.mu.Unlock()
	defer rf.persist()

	//fmt.Println("After locking, Server", rf.me, "is a", rf.state, "for command", command)
	if rf.state != Leader {
		isLeader = false
		return 0, 0, isLeader
	}
	term = rf.currentTerm
	rf.log = append(rf.log, LogEntry{
		Term:    term,
		Command: command,
	})
	go rf.Leader()

	rf.lastEntryIndex = len(rf.log) - 1
	index = rf.lastEntryIndex
	// fmt.Println("Server", rf.me, "now has lastEntryIndex:", rf.lastEntryIndex, "in the Start function")
	//rf.persist()
	return index, term, isLeader
}

// The tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. Your code can use killed() to
// check whether Kill() has been called. The use of atomic avoids the
// need for a lock.
//
// The issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. Any goroutine with a long-running loop
// should call killed() to check whether it should stop.
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func (rf *Raft) ticker() {
	// Your code here to check if a leader election should
	// be started and to randomize sleeping time using
	// time.Sleep().

	for !rf.killed() {
		rf.mu.Lock()
		state := rf.state
		rf.mu.Unlock()

		// Switch state {
		if state == Follower {
			ElectionTimeout := randTimeout()
			time.Sleep(ElectionTimeout)
			rf.mu.Lock()
			timeSince := rf.lastHeartbeat
			rf.matchIndex[rf.me] = rf.lastEntryIndex
			rf.nextIndex[rf.me] = rf.lastEntryIndex + 1
			rf.mu.Unlock()
			// fmt.Println("Server", rf.me, "is a Follower with last log index:", rf.matchIndex[rf.me], "and nextIndex:", rf.nextIndex[rf.me])

			// If timeout elaspses without recieving AppendEntries RPC, convert to Candidate
			if time.Since(timeSince) >= ElectionTimeout {
				rf.mu.Lock()
				rf.state = Candidate
				rf.currentTerm++ // Data race
				rf.votedFor = rf.me
				rf.persist()
				rf.mu.Unlock()

				// fmt.Println("Election started by server", rf.me, "on term", rf.currentTerm, "due to timeout and has voted for", rf.votedFor)
			}
		} else if state == Candidate {
			rf.mu.Lock()
			start := time.Now()
			rf.lastHeartbeat = time.Now()
			peers := rf.peers
			me := rf.me
			term := rf.currentTerm
			lastLogIndex := rf.lastEntryIndex
			lastLogTerm := rf.log[lastLogIndex].Term
			rf.mu.Unlock()

			total := len(peers)
			majority := (total / 2) + 1
			votes := 1

			// fmt.Println("Server", me, "is a candidate on term:", rf.currentTerm, "and has lastEntryIndex:", lastLogIndex)

			for peer := range peers {
				if peer == me {
					continue
				}

				// Start election by sending RequestVote RPCs to other servers
				go func(server int) {
					args := RequestVoteArgs{
						Term:         term,
						CandidateID:  me,
						LastLogIndex: lastLogIndex,
						LastLogTerm:  lastLogTerm,
					}
					reply := RequestVoteReply{}

					ok := rf.sendRequestVote(server, &args, &reply)

					rf.mu.Lock()
					defer rf.mu.Unlock()

					if !ok {
						// fmt.Println("Server", me, "failed to send RequestVote to server", server)
						return
					}
					if rf.currentTerm < reply.Term {
						rf.state = Follower
						rf.currentTerm = reply.Term
						rf.votedFor = -1
						rf.persist()
						// fmt.Println("Server", me, "is now a follower because of outdated term from RequestVote and has reset rf.voted for to:", rf.votedFor, "and rf.currentTerm to:", rf.currentTerm)
						return
					}
					if reply.VoteGranted {
						votes++
						if votes >= majority || time.Since(start) >= randTimeout() {
							return
						}
						time.Sleep(50 * time.Millisecond)
					} 
				}(peer)
				time.Sleep(50 * time.Millisecond)
			} 

			// After RPCs, check if...
			// A. The Candidate has recieved votes from a majority of servers
			// B. The election timeout has occurred

			rf.mu.Lock()

			if time.Since(start) >= randTimeout() {
				rf.state = Follower
				rf.votedFor = -1
				rf.persist()
				// fmt.Println("Server", me, "is now a follower because of election timeout and has reset rf.voted for to:", rf.votedFor)
				rf.mu.Unlock()
				return
			}

			if votes >= majority && rf.state == Candidate {
				// fmt.Println("Server", me, "has won the election on term:", term, "with,", votes, "votes")
				rf.state = Leader
				rf.lastHeartbeat = time.Now()
				rf.mu.Unlock()

				// Send empty AppendEntries RPCs to other servers
				for peer := range peers {
					if peer != me {
						go func(n int) {
							args := AppendEntriesArgs{
								Term:         term,
								LeaderID:     me,
								PrevLogIndex: lastLogIndex,
								PrevLogTerm:  lastLogTerm,
							}

							reply := AppendEntriesReply{}
							// fmt.Println("Server", me, "is sending initial heartbeats to server", n)
							ok := rf.SendAppendEntries(n, &args, &reply)

							if !ok {
								return
							}
							rf.mu.Lock()

							if reply.Term > rf.currentTerm {
								rf.state = Follower
								rf.currentTerm = reply.Term
								rf.votedFor = -1
								// fmt.Println("Server", me, "is now a follower because of outdated term from heartbeats and has reset rf.voted for to:", rf.votedFor)
								rf.persist()
								rf.mu.Unlock()
								return
							}
							rf.mu.Unlock()
						}(peer)
					}
				} 
			} else {
				rf.state = Follower
				rf.votedFor = -1
				rf.persist()
				rf.mu.Unlock()
				// fmt.Println("Server", me, "is now a follower because it didn't win the election and has reset rf.voted for to:", rf.votedFor)
			}
		} else if state == Leader {
			rf.Leader()
			// rf.mu.Lock()
			// me := rf.me

			// term := rf.currentTerm
			// log := rf.log
			// nodes := rf.peers
			// lastLogIndex := rf.lastEntryIndex // Index of last log entry

			// commitIndex := rf.commitIndex // Index of highest log entry known to be committed
			// highestIndex := rf.matchIndex // Index of highest log entry known to be replicated on server
			// nextIndex := rf.nextIndex

			// // This is only reset when the server becomes a leader
			// nextIndex[me] = lastLogIndex + 1
			// highestIndex[me] = lastLogIndex

			// rf.mu.Unlock()
			// // fmt.Println("Server", me, "is the leader on term:", term, "with lastLogIndex:", lastLogIndex)

			// // If there exists an N such that N > commitIndex, a majority
			// // of matchIndex[i] ≥ N, and log[N].term == currentTerm: set commitIndex = N
			// for N := commitIndex + 1; N <= lastLogIndex; N++ {
			// 	// fmt.Println("Entering loop with commitIndex:", commitIndex, "and N:", N, "lastLogIndex:", lastLogIndex)
			// 	count := 0
			// 	for x := 0; x < len(nodes); x++ {
			// 		if x == me {
			// 			continue
			// 		}
			// 		if highestIndex[x] >= N && log[N].Term == term { // Not entering this when follower disconnects
			// 			// fmt.Println("Highest index:", highestIndex[x])
			// 			count++
			// 			// fmt.Println("Count:", count)
			// 		}
			// 	}
			// 	if count >= (len(nodes) / 2) {
			// 		// Commit everything from the last commitIndex --> N since we have
			// 		// confirmed that a majority has replicated those logs
			// 		// fmt.Println("Majority reached!")
			// 		rf.mu.Lock()
			// 		// fmt.Println("CommitIndex before committing next entry:", rf.commitIndex)
			// 		i := rf.commitIndex + 1
			// 		for ; i <= N; i++ {
			// 			// fmt.Println("Committing index:", i, "with command")
			// 			rf.applyMsg <- ApplyMsg{
			// 				CommandValid: true,
			// 				Command:      log[i].Command,
			// 				CommandIndex: i,
			// 			}
			// 			rf.commitIndex = rf.commitIndex + 1
			// 			// fmt.Println("CommitIndex after committing entry:", rf.commitIndex)
			// 		}
			// 		rf.mu.Unlock()
			// 	}
			// }

			// for node := range nodes {
			// 	if node == me {
			// 		continue
			// 	}
			// 	rf.mu.Lock()
			// 	//highestIndex[node] = lastLogIndex
			// 	// Send appendEntries RPC
			// 	prevLogIndex := nextIndex[node] // Why is this resetting to 0 for Followers in TestPersist34B
			// 	// fmt.Println("Server", me, "is sending AppendEntries to server", node, "with term", term, "and prevLogIndex", prevLogIndex)
			// 	args := AppendEntriesArgs{
			// 		Term:         term,
			// 		LeaderID:     me,
			// 		PrevLogIndex: prevLogIndex,
			// 		PrevLogTerm:  rf.log[prevLogIndex].Term,
			// 		LeaderCommit: commitIndex,
			// 	}

			// 	// If last log index ≥ nextIndex for a follower:
			// 	// send AppendEntries RPC with log entries starting at nextIndex
			// 	if lastLogIndex >= prevLogIndex {
			// 		// fmt.Println("Server", rf.me, "is sending AppendEntries to server", node, "with term", term, "and log entries:", log[prevLogIndex + 1:])
			// 		args.Entries = rf.log[prevLogIndex + 1 : lastLogIndex + 1]
			// 	}

			// 	reply := AppendEntriesReply{}
			// 	rf.mu.Unlock()

			// 	go func(peer int) {
			// 		// fmt.Println("Server", me, "is sending AppendEntries", args.Entries, "to server", peer, "with term", term)
			// 		ok := rf.SendAppendEntries(peer, &args, &reply)

			// 		// If successful: update nextIndex and matchIndex for follower
			// 		// Elseif AppendEntries fails because of log inconsistency: decrement nextIndex and retry
			// 		if !ok {
			// 			// fmt.Println("Server", me, "failed to send AppendEntries to server", peer)
			// 			return
			// 		}
			// 		rf.mu.Lock()

			// 		if reply.Term > rf.currentTerm {
			// 			rf.state = Follower
			// 			rf.currentTerm = reply.Term // Don't increment term because outdated leader will think it's still the leader
			// 			rf.votedFor = -1
			// 			rf.persist()
			// 			rf.mu.Unlock()
			// 			// fmt.Println("Server", me, "is now a follower because of outdated term from AppendEntries and has reset rf.voted for to:", rf.votedFor)
			// 			return
			// 		}

			// 		if reply.Success {
			// 			// fmt.Println("Server", peer, "successfully appended entries")
			// 			rf.nextIndex[peer] = getMin(prevLogIndex + len(args.Entries), rf.lastEntryIndex + 1)
			// 			rf.matchIndex[peer] = prevLogIndex + len(args.Entries)
			// 			// fmt.Println("Now Server", peer, "'s nextIndex is: ", rf.nextIndex[peer])
			// 		} else {
			// 			if args.Term < reply.Term {
			// 				rf.state = Follower
			// 				rf.currentTerm = reply.Term
			// 				rf.votedFor = -1
			// 				rf.persist()
			// 				// fmt.Println("Server", me, "is now a follower because of outdated term from AppendEntries and has reset rf.voted for to:", rf.votedFor)
			// 				return
			// 			}
			// 			// Decrement next index and try again
			// 			if reply.ConflictTerm != -1 {
			// 				lastIndex := -1
			// 				for i := range rf.log {
			// 					if rf.log[i].Term == reply.ConflictTerm {
			// 						lastIndex = i
			// 						break
			// 					}
			// 				}
			// 				if lastIndex == -1 {
			// 					rf.nextIndex[peer] = reply.ConflictIndex
			// 				} else {
			// 					rf.nextIndex[peer] = lastIndex + 1
			// 				}
			// 			}
			// 		}
			// 		rf.mu.Unlock()
			// 	}(node)
			// 	time.Sleep(25 * time.Millisecond)
			// }
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (rf *Raft) Leader() {
	rf.mu.Lock()
			me := rf.me

			term := rf.currentTerm
			log := rf.log
			nodes := rf.peers
			lastLogIndex := rf.lastEntryIndex // Index of last log entry

			commitIndex := rf.commitIndex // Index of highest log entry known to be committed
			highestIndex := rf.matchIndex // Index of highest log entry known to be replicated on server
			nextIndex := rf.nextIndex

			// This is only reset when the server becomes a leader
			nextIndex[me] = lastLogIndex + 1
			highestIndex[me] = lastLogIndex

			rf.mu.Unlock()
			// fmt.Println("Server", me, "is the leader on term:", term, "with lastLogIndex:", lastLogIndex)

			// If there exists an N such that N > commitIndex, a majority
			// of matchIndex[i] ≥ N, and log[N].term == currentTerm: set commitIndex = N
			for N := commitIndex + 1; N <= lastLogIndex; N++ {
				// fmt.Println("Entering loop with commitIndex:", commitIndex, "and N:", N, "lastLogIndex:", lastLogIndex)
				count := 0
				for x := 0; x < len(nodes); x++ {
					if x == me {
						continue
					}
					if highestIndex[x] >= N && log[N].Term == term { // Not entering this when follower disconnects
						// fmt.Println("Highest index:", highestIndex[x])
						count++
						// fmt.Println("Count:", count)
					}
				}
				if count >= (len(nodes) / 2) {
					// Commit everything from the last commitIndex --> N since we have
					// confirmed that a majority has replicated those logs
					// fmt.Println("Majority reached!")
					rf.mu.Lock()
					// fmt.Println("CommitIndex before committing next entry:", rf.commitIndex)
					i := rf.commitIndex + 1
					for ; i <= N; i++ {
						// fmt.Println("Committing index:", i, "with command")
						rf.applyMsg <- ApplyMsg{
							CommandValid: true,
							Command:      log[i].Command,
							CommandIndex: i,
						}
						rf.commitIndex = rf.commitIndex + 1
						// fmt.Println("CommitIndex after committing entry:", rf.commitIndex)
					}
					rf.mu.Unlock()
				}
			}

			for node := range nodes {
				if node == me {
					continue
				}
				rf.mu.Lock()
				//highestIndex[node] = lastLogIndex
				// Send appendEntries RPC
				prevLogIndex := nextIndex[node] // Why is this resetting to 0 for Followers in TestPersist34B
				// fmt.Println("Server", me, "is sending AppendEntries to server", node, "with term", term, "and prevLogIndex", prevLogIndex)
				args := AppendEntriesArgs{
					Term:         term,
					LeaderID:     me,
					PrevLogIndex: prevLogIndex,
					PrevLogTerm:  rf.log[prevLogIndex].Term,
					LeaderCommit: commitIndex,
				}

				// If last log index ≥ nextIndex for a follower:
				// send AppendEntries RPC with log entries starting at nextIndex
				if lastLogIndex >= prevLogIndex {
					// fmt.Println("Server", rf.me, "is sending AppendEntries to server", node, "with term", term, "and log entries:", log[prevLogIndex + 1:])
					args.Entries = rf.log[prevLogIndex + 1 : lastLogIndex + 1]
				}

				reply := AppendEntriesReply{}
				rf.mu.Unlock()

				go func(peer int) {
					// fmt.Println("Server", me, "is sending AppendEntries", args.Entries, "to server", peer, "with term", term)
					ok := rf.SendAppendEntries(peer, &args, &reply)

					// If successful: update nextIndex and matchIndex for follower
					// Elseif AppendEntries fails because of log inconsistency: decrement nextIndex and retry
					if !ok {
						return
					}
					
					rf.mu.Lock()
					defer rf.mu.Unlock()
					
					if reply.Term > rf.currentTerm {
						rf.state = Follower
						rf.currentTerm = reply.Term // Don't increment term because outdated leader will think it's still the leader
						rf.votedFor = -1
						rf.persist()
						// rf.mu.Unlock()
						// fmt.Println("Server", me, "is now a follower because of outdated term from AppendEntries and has reset rf.voted for to:", rf.votedFor)
						return
					}

					if reply.Success {
						// fmt.Println("Server", peer, "successfully appended entries")
						rf.nextIndex[peer] = getMin(prevLogIndex + len(args.Entries), rf.lastEntryIndex + 1)
						rf.matchIndex[peer] = prevLogIndex + len(args.Entries)
						// fmt.Println("Now Server", peer, "'s nextIndex is: ", rf.nextIndex[peer])
					} else {
						if args.Term < reply.Term {
							rf.state = Follower
							rf.currentTerm = reply.Term
							rf.votedFor = -1
							rf.persist()
							// fmt.Println("Server", me, "is now a follower because of outdated term from AppendEntries and has reset rf.voted for to:", rf.votedFor)
							return
						}
						// Decrement next index and try again
						if reply.ConflictTerm != -1 {
							lastIndex := -1
							for i := range rf.log {
								if rf.log[i].Term == reply.ConflictTerm {
									lastIndex = i
									break
								}
							}
							if lastIndex == -1 {
								rf.nextIndex[peer] = reply.ConflictIndex
							} else {
								rf.nextIndex[peer] = lastIndex + 1
							}
						}
					}
					// rf.mu.Unlock()
				}(node)
				time.Sleep(25 * time.Millisecond)
			}
}

// The service or tester wants to create a Raft server. The ports
// of all the Raft servers (including this one) are in peers[]. This
// server's port is peers[me]. All the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}

	// Your initialization code here (4A, 4B).
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	rf.votedFor = -1
	rf.state = Follower
	rf.log = []LogEntry{
		{
			Term:    0,
			Command: nil,
		},
	}
	rf.nextIndex = make([]int, len(rf.peers))
	rf.matchIndex = make([]int, len(rf.peers))
	rf.applyMsg = applyCh

	// initialize from state persisted before a crash.
	rf.readPersist(persister.ReadRaftState())
	// fmt.Println("Server", me, "is a", rf.state, "has log:", rf.log, "and term:", rf.currentTerm, "and commit index:", rf.commitIndex, "and voted for", rf.votedFor)

	// Create a background goroutine that will kick off leader
	// election periodically by sending out `RequestVote` RPCs when it hasn't heard
	// from another peer for a while.
	go rf.ticker()

	return rf
}
