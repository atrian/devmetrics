package main

import "github.com/atrian/devmetrics/internal/agent"

func main() {
	statWatcher := agent.NewAgent()
	statWatcher.Run()
}
