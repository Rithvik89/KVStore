# Distributed In-Memory Key-Value Store

This is a distributed in-memory key-value store built in Go with leader election, 2-phase commit protocol, and ZooKeeper coordination. The system uses a master-follower architecture where leaders handle writes and followers handle reads.

Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.

## Working Effectively

### Bootstrap and Build
- Install dependencies:
  - `sudo apt-get update && sudo apt-get install -y openjdk-11-jdk zookeeper zookeeperd`
  - Go 1.23.3+ is required (Go 1.24.6 confirmed working)
- Build the application:
  - `cd /path/to/KVStore && go build ./cmd/http_api` -- takes ~0.4 seconds. NEVER CANCEL. Set timeout to 30+ minutes for safety.
- Run tests (no tests exist currently):
  - `go test ./...` -- takes ~0.2 seconds. NEVER CANCEL. Set timeout to 30+ minutes for safety.

### ZooKeeper Setup (CRITICAL)
ZooKeeper MUST be running and configured before starting the KVStore:
- Start ZooKeeper: `sudo systemctl start zookeeper`
- Verify ZooKeeper is running: `netstat -tuln | grep 2181`
- Create required ZooKeeper nodes:
  ```bash
  /usr/share/zookeeper/bin/zkCli.sh -server localhost:2181 create /election ""
  /usr/share/zookeeper/bin/zkCli.sh -server localhost:2181 create /workers ""
  /usr/share/zookeeper/bin/zkCli.sh -server localhost:2181 create /master ""
  /usr/share/zookeeper/bin/zkCli.sh -server localhost:2181 create /version "0"
  ```

### Run the Application
- Start as leader (first instance): `./http_api -port 8081`
- Start as follower (additional instances): `./http_api -port 8082`
- Application startup takes ~5 seconds. NEVER CANCEL during startup.
- Leader election happens automatically via ZooKeeper

## Validation Scenarios

### CRITICAL: Always test complete user scenarios after making changes
1. **Build and startup validation**:
   ```bash
   go build ./cmd/http_api
   ./http_api -port 8081 &
   # Wait for "KV Store is running on port: 8081" message
   ```

2. **Write operation validation (leader only)**:
   ```bash
   curl -X POST http://localhost:8081/api/v1/ \
     -H "Content-Type: application/json" \
     -d '{"key":"test", "value":"hello world"}'
   # Should return 200 OK or appropriate success response
   ```

3. **Multi-node validation** (optional):
   ```bash
   # Start second instance
   ./http_api -port 8082 &
   # Test read from follower
   curl -X GET http://localhost:8082/api/v1/ \
     -H "Content-Type: application/json" \
     -d '{"keys":["test"]}'
   ```

### Known Application Behavior
- **First instance becomes LEADER**: Can write, cannot read (returns "UnAuthorized action(GET) for a leader")
- **Subsequent instances become FOLLOWERS**: Can read, cannot write (returns "UnAuthorized action(POST) for a follower")
- Leader election uses ZooKeeper ephemeral sequential nodes
- WAL files (wal_[port].log) are created for transaction logging
- **WAL conflict detection**: After the first successful write, subsequent writes may show "Failed to write to WAL" due to conflict detection - this is normal distributed systems behavior

## Common Issues and Solutions

### Build Issues
- If build fails with "too many arguments" for ReplicationManager: This is fixed in the codebase
- If cluster manager hangs during startup: This is fixed in the codebase (blocking loop moved to goroutine)

### Runtime Issues  
- **"panic: zk: node does not exist"**: Run ZooKeeper setup commands above
- **"zk: could not connect to a server"**: Start ZooKeeper with `sudo systemctl start zookeeper`
- **"Failed to write to WAL"**: Ensure /version node exists in ZooKeeper with proper JSON integer format
- **WAL conflict detection**: Normal behavior - indicates distributed coordination is working

### Expected Startup Messages
```
Election Manager initialized
connected to [::1]:2181 or 127.0.0.1:2181
This instance is the leader (or "not the leader")
Registered master: /master/... (or "Registered worker: /workers/...")
WAL Manager initialized
Replication Manager initialized  
KV Store is running on port: [PORT]
```

## Key Project Components

### Repository Structure
- `cmd/http_api/` - Main HTTP API server and entry point
- `internal/cluster/` - Cluster management and quorum logic
- `internal/elections/` - Leader election using ZooKeeper
- `internal/kv/` - In-memory key-value store implementation
- `internal/replication/` - 2-phase commit replication to followers
- `internal/wal/` - Write-Ahead Log for durability
- `pkg/kvstore/` - Public interfaces for KV operations
- `utils/` - HTTP utility functions

### API Endpoints
- `GET /api/v1/` - Read keys (followers only)  
  - Request: `{"keys":["key1","key2"]}`
  - Response: `["value1","value2"]`
- `POST /api/v1/` - Write key-value pair (leaders only)
  - Request: `{"key":"mykey", "value":"myvalue"}`

### Architecture Patterns
- **Leader-Follower**: Distributed via ZooKeeper coordination
- **2-Phase Commit**: For consistent writes across cluster
- **WAL**: Write-Ahead Logging for crash recovery
- **Quorum-based**: Configurable read/write quorums

### Important Files to Check After Changes
- Always rebuild after modifying any Go files
- Check WAL files (wal_*.log) after write operations
- Monitor ZooKeeper nodes: `/election`, `/workers`, `/master`, `/version`
- Verify HTTP endpoints are responding on expected ports

### Development Workflow
- **ALWAYS** rebuild and test the complete application after any code changes
- **NEVER** skip the ZooKeeper setup - the application will panic without it  
- **ALWAYS** test both leader and follower behavior when applicable
- WAL files contain transaction history - check them for debugging

## Example Session Output
```
ls -la [repo-root]:
cmd/
go.mod
go.sum
internal/
pkg/
readme.md
utils/

go build ./cmd/http_api output:
(builds in ~0.4 seconds with no output on success)

./http_api -port 8081 output:
Election Manager initialized
This instance is the leader
WAL Manager initialized
Replication Manager initialized
KV Store is running on port: 8081
```