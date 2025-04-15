package cluster

import (
	"fmt"

	"github.com/go-zookeeper/zk"
)

type ClusterManager struct {
	KvPort      int      `json:"kv_port"`
	ZkClient    *zk.Conn `json:"zk_client"`
	ClusterSize int32    `json:"cluster_size"`
	WriteQuorum int32    `json:"write_quorum"`
	ReadQuorum  int32    `json:"read_quorum"`
}

func NewClusterManager(kv_port int, zkClient *zk.Conn) *ClusterManager {
	return &ClusterManager{
		KvPort:   kv_port,
		ZkClient: zkClient,
	}
}

func (cm *ClusterManager) getClusterSize() int32 {
	// Get the cluster size from Zookeepr
	path := "/workers" // Zookeeper path for leader election
	// Get the list of children nodes
	children, _, err := cm.ZkClient.Get(path)
	if err != nil {
		panic(err)
	}

	return int32(len(children))
}

func (cm *ClusterManager) getWriteQuorum() int32 {
	// Calculate the write quorum
	return (cm.ClusterSize / 2) + 1
}

func (cm *ClusterManager) getReadQuorum() int32 {
	// Calculate the read quorum
	return (cm.ClusterSize / 2) + 1
}

func (cm *ClusterManager) InitializeClusterMetadata() {
	path := "/workers" // Zookeeper path for leader election

	updateClusterDetails := func() {
		cm.ClusterSize = cm.getClusterSize()
		cm.WriteQuorum = cm.getWriteQuorum()
		cm.ReadQuorum = cm.getReadQuorum()
	}

	updateClusterDetails()

	// Get the list of children nodes
	// here we are watching for changes in cluster size like if some replicas are added or removed/crashed

	_, _, ch, err := cm.ZkClient.GetW(path)
	if err != nil {
		panic(err)
	}

	for ev := range ch {
		if ev.Type == zk.EventNodeChildrenChanged {
			fmt.Println("Cluster size changed, resetting cluster details")
			updateClusterDetails()
		}
	}
}
