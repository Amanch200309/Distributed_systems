# Chord DHT - Testing Guide

## Build Instructions

To build the Chord application:

```bash
cd /home/d4n3r/Desktop/Distributed_systems/LAB3
go build -o chord-dht ./app/main.go
```

This will create an executable named `chord-dht` in the current directory.

## Running Chord Nodes

### Starting the First Node (Create New Ring)

To create a new Chord ring:

```bash
./chord-dht -a 127.0.0.1 -p 8001 --ts 3000 --tff 1000 --tcp 3000 -r 4
```

Parameters:
- `-a 127.0.0.1`: IP address to bind to
- `-p 8001`: Port to listen on
- `--ts 3000`: Stabilize every 3000ms (3 seconds)
- `--tff 1000`: Fix fingers every 1000ms (1 second)
- `--tcp 3000`: Check predecessor every 3000ms (3 seconds)
- `-r 4`: Maintain 4 successors

### Joining Additional Nodes to the Ring

Open new terminal windows and run:

**Node 2:**
```bash
./chord-dht -a 127.0.0.1 -p 8002 --ja 127.0.0.1 --jp 8001 --ts 3000 --tff 1000 --tcp 3000 -r 4
```

**Node 3:**
```bash
./chord-dht -a 127.0.0.1 -p 8003 --ja 127.0.0.1 --jp 8001 --ts 3000 --tff 1000 --tcp 3000 -r 4
```

**Node 4:**
```bash
./chord-dht -a 127.0.0.1 -p 8004 --ja 127.0.0.1 --jp 8002 --ts 3000 --tff 1000 --tcp 3000 -r 4
```

Parameters for joining:
- `--ja 127.0.0.1`: Join address (IP of existing node)
- `--jp 8001`: Join port (port of existing node)

## Testing Commands

After starting nodes, you can test the following commands:

### 1. PrintState

View the current state of a node:

```
> PrintState
```

This displays:
- Node's own ID and address
- Stored files
- Predecessor information
- Successor list
- Finger table

### 2. StoreFile

Create a test file first:

```bash
echo 'Hello from Chord DHT!' > /tmp/test.txt
echo 'This is another test file.' > /tmp/test2.txt
echo 'Distributed systems are cool!' > /tmp/test3.txt
```

Then store files in the ring:

```
> StoreFile /tmp/test.txt test.txt
> StoreFile /tmp/test2.txt test2.txt
> StoreFile /tmp/test3.txt test3.txt
```

### 3. Lookup

Look up a file in the ring:

```
> Lookup test.txt
> Lookup test2.txt
> Lookup test3.txt
```

This will show:
- Which node owns the file
- The node's ID and address
- The file content

## Complete Test Scenario

### Step 1: Start Multiple Nodes

Open 3 terminal windows:

**Terminal 1 (Node 1):**
```bash
./chord-dht -a 127.0.0.1 -p 8001 --ts 3000 --tff 1000 --tcp 3000 -r 4
```

Wait a few seconds, then:

**Terminal 2 (Node 2):**
```bash
./chord-dht -a 127.0.0.1 -p 8002 --ja 127.0.0.1 --jp 8001 --ts 3000 --tff 1000 --tcp 3000 -r 4
```

Wait a few seconds, then:

**Terminal 3 (Node 3):**
```bash
./chord-dht -a 127.0.0.1 -p 8003 --ja 127.0.0.1 --jp 8001 --ts 3000 --tff 1000 --tcp 3000 -r 4
```

### Step 2: Wait for Stabilization

Wait about 10-15 seconds for the stabilization routines to update finger tables and successor lists.

### Step 3: Check Node State

In each terminal, run:
```
> PrintState
```

You should see:
- Each node has different successors
- Finger tables point to different nodes
- Predecessor information is populated

### Step 4: Store Files

Create test files:
```bash
echo 'File 1 content' > /tmp/file1.txt
echo 'File 2 content' > /tmp/file2.txt
echo 'File 3 content' > /tmp/file3.txt
```

In any node's terminal, store files:
```
> StoreFile /tmp/file1.txt file1.txt
> StoreFile /tmp/file2.txt file2.txt
> StoreFile /tmp/file3.txt file3.txt
```

### Step 5: Verify Distribution

Run `PrintState` on each node to see which files are stored where. Files should be distributed based on the hash of their names.

### Step 6: Lookup Files

From any node, look up the files:
```
> Lookup file1.txt
> Lookup file2.txt
> Lookup file3.txt
```

The lookup should:
1. Hash the filename
2. Find the correct node (successor of the hash)
3. Retrieve and display the file content

### Step 7: Test from Different Nodes

Try the lookups from different nodes to verify that Chord routing works correctly regardless of which node initiates the request.

## Expected Behavior

1. **Ring Formation**: Nodes should form a ring where each node points to its successor
2. **Finger Tables**: Should be populated with shortcuts to distant nodes
3. **File Storage**: Files are stored at the successor of hash(filename)
4. **Lookup**: Any node can find any file by following finger table pointers
5. **Stabilization**: New nodes integrate into the ring within several stabilization cycles

## Troubleshooting

### Problem: Nodes can't connect
- Check that ports are not already in use: `netstat -tuln | grep 800`
- Ensure firewall allows local connections

### Problem: Finger tables not updating
- Wait longer (10-15 seconds) for stabilization
- Check that --tff is set (e.g., 1000ms)

### Problem: Files not found
- Verify the file was stored successfully (check output)
- Run `PrintState` to see which node actually has the file
- Make sure nodes have stabilized before testing lookups

### Problem: Build fails
- Ensure Go is installed: `go version`
- Check that you're in the correct directory
- Run `go mod tidy` to fix dependencies

## Advanced Testing

### Test with More Nodes

Start 5-10 nodes to see better distribution:

```bash
for i in {1..10}; do
  port=$((8000 + i))
  if [ $i -eq 1 ]; then
    ./chord-dht -a 127.0.0.1 -p $port --ts 3000 --tff 1000 --tcp 3000 -r 4 &
  else
    sleep 2
    ./chord-dht -a 127.0.0.1 -p $port --ja 127.0.0.1 --jp 8001 --ts 3000 --tff 1000 --tcp 3000 -r 4 &
  fi
done
```

### Test Node Failure

1. Start 4 nodes
2. Store some files
3. Kill one node (Ctrl+C)
4. Wait for stabilization
5. Try to lookup files - some may be lost (this is expected in basic Chord without replication)

### Performance Testing

Store many files and measure lookup time:

```bash
for i in {1..100}; do
  echo "Content $i" > /tmp/file$i.txt
done
```

Then store them all and measure lookup performance.

## Exit

To exit a node, type:
```
> quit
```

Or press Ctrl+C.

## Notes

- The basic implementation doesn't include data migration when nodes join/leave
- Files may be lost if the node storing them fails (no replication yet)
- For production use, implement successor list replication (advanced part)
- The -i flag for ID override is not fully implemented in this basic version
