package main

import (
	"rps_main/internal/server"
)

func main() {
	s, _ := server.NewServer(":9000")
	s.Start()
}
