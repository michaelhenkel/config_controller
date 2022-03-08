package main

import (
	"github.com/michaelhenkel/config_controller/pkg/k8s"
	"github.com/michaelhenkel/config_controller/pkg/server"

	"github.com/michaelhenkel/config_controller/pkg/db"
)

func main() {
	var stopCh = make(chan struct{})
	dbClient := db.NewClient()
	serverClient := server.NewClient()
	k8sClient := k8s.NewClient(dbClient)
	go dbClient.Start()
	go k8sClient.Start()
	go serverClient.Start(k8sClient)
	<-stopCh
}
