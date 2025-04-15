# 🗃️ Distributed Key-Value Store

A distributed in-memory key-value store. It supports **master-based writes**, **read/write quorums**, **write-ahead logging**, and **read replicas** to ensure durability and availability.

## ✨ Features

- 🧠 **Strong Consistency** via master-coordinated write quorum.
- 🧾 **Write-Ahead Log (WAL)** to ensure durability and replay on recovery.
- 🔁 **Read Replicas** for load-balanced and quorum-based reads.
- 🔒 **Conflict Resolution** using log index versioning and resync mechanism.
- ⚙️  **Quorum Configurable**: Tunable read (`R`) and write (`W`) quorum settings.
- 🧠 **Background Snapshotting** to persist the current in-memory state.
- 🔄 **Replica Sync Jobs** to identify and heal WAL gaps and drift.
- 📊 **In-Memory Store** for low-latency reads and writes.
- 💥 **Fault Tolerance**: Supports replica resyncs and partial failures.





