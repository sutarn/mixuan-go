package main

import (
	"github.com/coraldane/mixuan-go/mixuan"
	"log"
	"os"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	logFileWriter, _ := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	log.SetOutput(logFileWriter)

	server := mixuan.NewTcpServer(1200)
	err := server.Start()
	if nil != err {
		log.Fatalf("Server start at port 1200 fail, error is %#v.", err)
	}
}
