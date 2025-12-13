# Lab 3: Chord DHT Implementation - Presentation

## Overview
We implemented a Chord Distributed Hash Table (DHT) in Go for distributed file storage. The system uses SHA-1 hashing (160-bit identifier space) and supports dynamic node joining/leaving with automatic stabilization.

## Key Components

### 1. Node Structure (`node.go`)
- **Node struct**: Maintains ID, predecessor, successor list (r successors), and finger table (m entries)
- **Join()**: Contacts bootstrap node, finds initial successor via iterative lookup, sets predecessor to nil
- **Create()**: Initializes new ring with all pointers to self

### 2. Iterative Lookup (`lookup.go`)
- **Find()**: Implements iterative lookup (not recursive) - starts from local node, repeatedly calls FindSuccessorRPC until successor found
- **findSuccessor()**: Single-step lookup - returns successor if ID falls in range, otherwise returns closest preceding node
- **closestPrecedingNode()**: Searches finger table backwards to find closest node preceding target ID
- **between()**: Handles circular identifier space arithmetic with wraparound

### 3. Stabilization (`stabilize.go`)
- **stabilize()**: Verifies successor, checks successor's predecessor, updates if needed, notifies successor, copies successor list
- **fixFinger()**: Updates one finger table entry per invocation (round-robin), computes start = (n + 2^i) mod 2^m
- **checkPredecessor()**: Pings predecessor, sets to nil if dead
- **RunMaintenance()**: Spawns goroutines running stabilize/fixFinger/checkPredecessor at specified intervals

### 4. File Operations (`storefile.go`)
- **StoreFile()**: Reads file from disk, hashes filename (not content!), finds responsible node via Find(), sends file via RPC
- **Lookup()**: Hashes filename, performs lookup, retrieves file contents via GetFileRPC
- Files stored in both `Data` map (hash→bytes) and `Files` map (hash→metadata with filename)

### 5. RPC Communication (`rpc.go`)
- **TCP-based RPC**: All inter-node communication via Go's net/rpc package
- **Key RPCs**: FindRPC (complete lookup), FindSuccessorRPC (single step), GetPredecessorRPC, NotifyRPC, PingRPC, StoreFileRPC, GetFileRPC, GetSuccessorListRPC
- **call()**: Helper wrapping RPC with automatic connection management

### 6. Monitoring (`monitoring.go`)
- **PrintState()**: Shows local node info, successor list, finger table, predecessor, stored files
- **PrintRingStatus()**: Crawls entire ring, displays all nodes with their connections and files
- **PrintFileDistribution()**: Table view of which files are on which nodes
- **VerifyFileOwnership()**: Validates files are stored on correct nodes based on hash

## Problem Solutions

### Problem: Circular Identifier Space
**Solution**: `between()` function handles wraparound - if start > end, checks if id > start OR id < end

### Problem: Node Failures
**Solution**: 
- Successor list (r successors) provides backup when primary successor fails
- stabilize() shifts successor list left on failure
- checkPredecessor() detects dead predecessors
- Periodic maintenance ensures eventual consistency

### Problem: Concurrent Access
**Solution**: RWMutex protects Node struct, read locks for queries, write locks for modifications

### Problem: Iterative vs Recursive Lookup
**Solution**: Implemented iterative approach (Find()) - no blocking, each RPC returns immediately, client drives lookup

### Problem: Finger Table Maintenance
**Solution**: fixFinger() updates one entry at a time (index cycles 0→m-1), spread over time controlled by --tff parameter

### Problem: Ring Convergence
**Solution**: 
- notify() allows nodes to inform successors of their existence
- stabilize() propagates successor list through ring
- Combined with periodic execution ensures ring converges after joins

## Command-Line Interface
```bash
# Create new ring
./main -a 127.0.0.1 -p 8000 --ts 1000 --tff 500 --tcp 1000 -r 3

# Join existing ring  
./main -a 127.0.0.1 -p 8001 --ja 127.0.0.1 --jp 8000 --ts 1000 --tff 500 --tcp 1000 -r 3

# Commands
> storefile <path>    # Store file in DHT
> lookup <filename>   # Find file owner
> printstate          # Show local state
> ring                # Visualize full ring
> files               # File distribution
```

## Testing
- Successfully tested with 8+ nodes
- Files correctly distributed based on hash
- Ring maintains consistency through joins
- Successor list provides fault tolerance
