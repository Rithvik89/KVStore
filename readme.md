# 🚀 KV Store – v1.0 Release Changelog

## ✅ Core Features

### 🧠 Master-Slave Architecture
- Master node handles `/put` and `/delete`
- Followers handle `/get`

### 🔐 Strong Consistency
- Achieved using **quorum-based reads and writes**

### 🔄 gRPC-Based Replication
- Master node sends WAL entries to followers via gRPC for writes

### 🎯 Load Balancing
- Load balancer routes:
  - **Writes** to master
  - **Reads** via **round-robin** among replicas

### ⚡ On-the-Fly Replica Addition
- Add new replicas at runtime with **no downtime**

### 📝 Write-Ahead Log (WAL)
- All writes go to a persistent WAL before application
- Guarantees durability and crash recovery

### 🔄 WAL Version Sync
- Followers detect WAL mismatches
- Call a **sync API** on master to reconcile
