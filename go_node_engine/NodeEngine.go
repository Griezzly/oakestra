package main

import (
	"flag"
	"go_node_engine/containers"
	"go_node_engine/interfaces"
	"go_node_engine/jobs"
	"go_node_engine/model"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var clusterAddress = flag.String("a", "localhost", "Address of the cluster orchestrator without port")
var clusterPort = flag.String("p", "10000", "Port of the cluster orchestrator")

func main() {
	flag.Parse()

	//connect to container runtime
	runtime := containers.GetContainerdClient()
	defer runtime.StopContainerdClient()

	//hadshake with the cluster orchestrator to get mqtt port and node id
	handshakeResult := clusterHandshake()

	//binding the node MQTT client
	interfaces.InitMqtt(handshakeResult.NodeId, *clusterAddress, handshakeResult.MqttPort)

	//starting node status background job. One udpate every 30 seconds
	go jobs.NodeStatusUpdater(time.Second * 10)
	//TODO: start tasks monitoring job

	// catch SIGETRM or SIGINTERRUPT
	termination := make(chan os.Signal, 1)
	signal.Notify(termination, syscall.SIGTERM, syscall.SIGINT)
	select {
	case ossignal := <-termination:
		log.Printf("Terminating the NodeEngine, signal:%v", ossignal)
	}
}

func clusterHandshake() interfaces.HandshakeAnswer {
	log.Printf("INIT: Starting handshake with cluster orhcestrator %s:%s", *clusterAddress, *clusterPort)
	node := model.GetNodeInfo()
	log.Printf("Node Statistics: \n__________________")
	log.Printf("CPU Cores: %d", node.CpuCores)
	log.Printf("CPU Usage: %f", node.CpuUsage)
	log.Printf("Mem Usage: %f", node.MemoryUsed)
	log.Printf("GPU Present: %t", len(node.GpuInfo) > 0)
	log.Printf("\n________________")
	clusterReponse := interfaces.ClusterHandshake(*clusterAddress, *clusterPort)
	log.Printf("Got cluster response with MQTT port %s and node ID %s", clusterReponse.MqttPort, clusterReponse.NodeId)
	node.SetNodeId(clusterReponse.NodeId)
	return clusterReponse
}
