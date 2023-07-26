package main

import (
	"log"
	"net"
	"os"
)

var strippedframe []byte

func main() {
	// listen to incoming udp packets
	udpServer, err := net.ListenPacket("udp", ":1053")
	if err != nil {
		log.Fatal(err)
	}
	defer udpServer.Close()

	// myfile, err := os.Create("/tmp/testvideoraw.mp4")
	// checkErr(err)
	for {
		buf := make([]byte, 65536, 65536)
		_, _, err := udpServer.ReadFrom(buf)
		if err != nil {
			continue
		}
		for j := 65536; j != 0; j-- {
			if buf[j-1] != 0 {
				strippedframe = buf[:j]
				break
			}
		}

		os.Stdout.Write(strippedframe)
	}

}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
