# Chord DHT - Quick Start Guide

## Build

```bash
cd /home/d4n3r/Desktop/Distributed_systems/LAB3
go build -o chord-dht ./app/main.go
```

## Quick Test (3 Nodes)

### Terminal 1 - Start First Node
```bash
./chord-dht -a 127.0.0.1 -p 8001 --ts 3000 --tff 1000 --tcp 3000 -r 4
```

### Terminal 2 - Join Second Node (wait 3 seconds first)
```bash
./chord-dht -a 127.0.0.1 -p 8002 --ja 127.0.0.1 --jp 8001 --ts 3000 --tff 1000 --tcp 3000 -r 4
```

### Terminal 3 - Join Third Node (wait 3 seconds first)
```bash
./chord-dht -a 127.0.0.1 -p 8003 --ja 127.0.0.1 --jp 8001 --ts 3000 --tff 1000 --tcp 3000 -r 4
```

## Test Commands

### 1. Check Node State
In any terminal:
```
> PrintState
```

### 2. Store Files
First, create test files:
```bash
echo 'Hello World' > /tmp/hello.txt
echo 'Test data' > /tmp/test.txt
```

Then in any node terminal:
```
> StoreFile /tmp/hello.txt hello.txt
> StoreFile /tmp/test.txt test.txt
```

### 3. Lookup Files
From any node:
```
> Lookup hello.txt
> Lookup test.txt
```

### 4. Verify Distribution
Run `PrintState` on each node to see which files are stored where.

## Command Reference

| Command | Usage | Description |
|---------|-------|-------------|
| `Lookup` | `Lookup <filename>` | Find and display a file from the ring |
| `StoreFile` | `StoreFile <filepath> <filename>` | Store a local file in the ring |
| `PrintState` | `PrintState` | Display current node state |
| `quit` | `quit` or `exit` | Exit the node |

## CLI Arguments

### Required
- `-a <ip>` - IP address to bind to
- `-p <port>` - Port to listen on
- `--ts <ms>` - Stabilize interval (1-60000ms)
- `--tff <ms>` - Fix fingers interval (1-60000ms)
- `--tcp <ms>` - Check predecessor interval (1-60000ms)
- `-r <num>` - Number of successors (1-32)

### Optional (for joining)
- `--ja <ip>` - IP of node to join
- `--jp <port>` - Port of node to join

### Optional
- `-i <id>` - Override node ID (40 hex chars)

## Tips

1. **Wait for stabilization**: After starting nodes, wait 5-10 seconds before testing
2. **Check state regularly**: Use `PrintState` to monitor finger table updates
3. **Test from different nodes**: Verify lookups work from any node
4. **Multiple files**: Store several files to see distribution across nodes

## Troubleshooting

**Can't connect to node:**
- Make sure the bootstrap node is running
- Check ports are available: `netstat -tuln | grep 800`

**Files not found:**
- Wait longer for stabilization
- Check which node has the file with `PrintState`

**Build errors:**
- Run `go mod tidy`
- Ensure you're in the LAB3 directory

## See Also

- `TESTING.md` - Comprehensive testing guide
- `IMPLEMENTATION_SUMMARY.md` - Details of implementation
- `README.md` - Project overview
