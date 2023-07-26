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
		log.Fatal(err)
	}
}

var msg1, msg2 unix.Msghdr
var iov [2]unix.Iovec

func main() {
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

	// myfile, err := os.Create("/mnt/sdcard/testvideoraw.mp4")
	// checkErr(err)
	log.Print("start")

	packagelen := 0
	for {
		_, _, err = unix.Syscall(unix.SYS_RECVMSG, uintptr(sockFd), uintptr(unsafe.Pointer(&msg2)), uintptr(syscall.MSG_NOSIGNAL))
		if err == syscall.EAGAIN {
			_, err = syscall.EpollWait(epollfd, []syscall.EpollEvent{socketEpoll}, 59743)
			checkErr(err)
		} else {
			header, err := syscall.Mmap(int(controlmsg[12]), 0, 10, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_PRIVATE)
			checkErr(err)
			packagelen = int(binary.LittleEndian.Uint16(header[8:10]))
			err = syscall.Munmap(header)
			checkErr(err)
			videoaddr, err := syscall.Mmap(int(controlmsg[12]), 0, packagelen+96, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_PRIVATE)
			checkErr(err)

			_, err = conn.Write(videoaddr[96:])
			if err != nil {
				println("Write data failed:", err.Error())
				continue
			}

			// _, err = myfile.Write(videoaddr)
			// checkErr(err)
			err = syscall.Munmap(videoaddr)
			checkErr(err)
			err = syscall.Close(int(controlmsg[12]))
			checkErr(err)
		}
	}
	log.Print("stop")
}
