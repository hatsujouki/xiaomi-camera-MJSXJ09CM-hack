package main

import (
	"log"
	"net"
	"os"
)

var strippedframe []byte
var framesize int

func main() {
	udpServer, err := net.ListenPacket("udp", ":1053")
	if err != nil {
		log.Fatal(err)
	}
	defer udpServer.Close()

	// myfile, err := os.Create("/tmp/tempvideo.mp4")
	// checkErr(err)
	for {
		buf := make([]byte, 4096, 4096)
		_, _, err := udpServer.ReadFrom(buf)
		if err != nil {
			continue
		}
		if Equal(buf[:4], []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
			for j := 4096; j != 0; j-- {
				if buf[j-1] != 0 {
					// myfile.Write(buf[96:j])
					os.Stdout.Write(buf[96:j])
					break
				}
			}

		} else {
			for j := 4096; j != 0; j-- {
				if buf[j-1] != 0 {
					// myfile.Write(buf[:j])
					os.Stdout.Write(buf[:j])
					break
				}
			}
		}
	}

}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func Equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

