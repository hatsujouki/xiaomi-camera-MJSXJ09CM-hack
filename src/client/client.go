package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

func checkErr(err error) {
	if err != nil {
		log.Fatal("Stopped due to checkErr\n", err)
	}
}

var msg1, msg2 unix.Msghdr
var iov [2]unix.Iovec

func main() {
	// IDR_W_RADL := 0x26
	// TRAIL_R := 0x02

	if len(os.Args) < 2 || os.Args[1] == "" {
		fmt.Println("Use ./client 192.168.1.65:1053")
		os.Exit(1)
	}
	serverForStream := os.Args[1]
	videoSocket := syscall.SockaddrUnix{Name: "/run/video_mainstream/control"}

	udpServer, err := net.ResolveUDPAddr("udp", serverForStream)
	checkErr(err)
	conn, err := net.DialUDP("udp", nil, udpServer)
	checkErr(err)
	defer conn.Close()

	sockFd, err := syscall.Socket(syscall.AF_LOCAL, syscall.SOCK_SEQPACKET|syscall.SOCK_CLOEXEC|syscall.SOCK_NONBLOCK, 0)
	checkErr(err)
	defer syscall.Close(sockFd)
	err = syscall.Connect(sockFd, &videoSocket)
	checkErr(err)

	iovepart1 := make([]byte, 4, 4)
	iovepart2 := make([]byte, 65536, 65536)

	iov[0].Base = &iovepart1[0]
	iov[0].SetLen(len(iovepart1))
	iov[1].Base = &iovepart2[0]
	iov[1].SetLen(len(iovepart2))

	msg1.Name = nil
	msg1.Namelen = 0
	msg1.Iov = &iov[:][0]
	msg1.SetIovlen(len(iov))
	msg1.Control = nil
	msg1.Controllen = 0
	msg1.Flags = 0

	controlmsg := make([]byte, 16, 16)

	msg2.Name = nil
	msg2.Namelen = 0
	msg2.Iov = &iov[:][0]
	msg2.SetIovlen(len(iov))
	msg2.Control = &controlmsg[0]
	msg2.Controllen = 16
	msg2.Flags = 0

	_, _, _ = unix.Syscall(unix.SYS_RECVMSG, uintptr(sockFd), uintptr(unsafe.Pointer(&msg1)), uintptr(syscall.MSG_NOSIGNAL))
	socketEpoll := syscall.EpollEvent{Events: unix.EPOLLIN, Fd: int32(sockFd), Pad: int32(1)}
	epollfd, err := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	checkErr(err)

	err = syscall.EpollCtl(epollfd, syscall.EPOLL_CTL_ADD, sockFd, &socketEpoll)
	checkErr(err)

	log.Print("start")

	frameLenLittle := 0

	for {
		_, _, err = unix.Syscall(unix.SYS_RECVMSG, uintptr(sockFd), uintptr(unsafe.Pointer(&msg2)), uintptr(syscall.MSG_NOSIGNAL))
		if err == syscall.EAGAIN {
			_, err = syscall.EpollWait(epollfd, []syscall.EpollEvent{socketEpoll}, 59743)
			checkErr(err)
			// ниже я пытался решить проблему с тем, что в случайные моменты времени client убивается прерыванием. Камера начинает виснуть, как решить пока не понятно.
			/*   _, err := syscall.EpollWait(epollfd, []syscall.EpollEvent{socketEpoll}, 59743)
			     if err == syscall.EINTR {
			                                  fmt.Println("interrupted")
			                                  continue
			                          } else if err != nil {
			                                  log.Fatal(err)
			     }
			*/
			/*   for {
			      _, err := syscall.EpollWait(epollfd, []syscall.EpollEvent{socketEpoll}, 59743)
			      if err == syscall.EINTR {
			       fmt.Println("interrupted")
			       continue
			      } else if err != nil {
			       log.Fatal(err)
			      } else {
			       break
			      }
			     }
			*/
		} else {
			frameLenLittle = int(binary.LittleEndian.Uint32(iovepart2[4:8]))
			videoaddr, err := syscall.Mmap(int(controlmsg[12]), 0, frameLenLittle, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_PRIVATE)
			checkErr(err)

			/*   if Equal(videoaddr[100:102], []byte{0x40, 0x01}){
			      if videoaddr[185] == byte(IDR_W_RADL) {
			       fmt.Println("IDR_W_RADL")
			      }
			     } else {
			      if Equal(videoaddr[100:102], []byte{byte(TRAIL_R), 0x01}){
			       fmt.Println("TRAIL_R")
			      }
			     }
			*/
			for i := 0; i < frameLenLittle; i += 4096 {
				_, err = conn.Write(videoaddr[i : i+4096])
				if err != nil {
					println("Write data failed:", err.Error())
					continue
				}
			}

			err = syscall.Munmap(videoaddr)
			checkErr(err)
			err = syscall.Close(int(controlmsg[12]))
			checkErr(err)
		}
	}
	log.Print("stop")
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

