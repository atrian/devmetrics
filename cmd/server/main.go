package main

import "github.com/atrian/devmetrics/internal/server"

func main() {
	statServer := server.NewServer()
	statServer.Run()
}
