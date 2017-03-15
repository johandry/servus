package main

import (
	"log"
	"os"

	"github.com/hashicorp/serf/serf"
)

func setupCluster(advertiseAddr string, clusterAddr string) (*serf.Serf, error) {
	conf := serf.DefaultConfig()
	conf.Init()
	conf.MemberlistConfig.AdvertiseAddr = advertiseAddr

	cluster, err := serf.Create(conf)
	if err != nil {
		return nil, err
	}

	_, err = cluster.Join([]string{clusterAddr}, true)
	if err != nil {
		log.Printf("Could not join to the cluster at %s, creating a new one. %v\n", clusterAddr, err)
	}

	return cluster, nil
}

func main() {
	cluster, err := setupCluster(os.Getenv("SERVUS_ADVERTISE_ADDRESS"), os.Getenv("SERVUS_CLUSTER_ADDRESS"))
	if err != nil {
		log.Fatalln(err)
	}
	defer cluster.Leave()
}
