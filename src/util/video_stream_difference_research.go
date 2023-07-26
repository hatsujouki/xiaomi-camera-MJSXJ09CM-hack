package main

import (
	"fmt"
	//"io"
	//"log"
	"os"
	//"time"
	"encoding/binary"
)

var strippedframe []byte

func main() {
	myfile1, _ := os.ReadFile("/home/reg/Desktop/развлечения/разборки с камерой/полученный с сокета поток/compared_part_my_own_binary.mp4")
	myfile2, _ := os.ReadFile("/home/reg/Desktop/развлечения/разборки с камерой/полученный с сокета поток/compared_part_miio_record.mp4")
	framenumber := 0
	laststop := 0
	for i := 0; i < len(myfile1); i += 4096 {
		currentpart := myfile1[i : i+4096]
		//current_metadata := currentpart[:100]
		currentframe := currentpart[100:]
		fmt.Println("\nframe number: ", framenumber)
		//fmt.Println(current_metadata)
		for j := len(currentframe); j != 0; j-- {
			if currentframe[j-1] != 0 {
				strippedframe = currentframe[:j]
				fmt.Println("frame length in 1st file: ", len(currentframe[:j]))
				break
			}
		}
		f2pos := 0
		frameLenInSecondFile := 0

		found := false
		for (f2pos+len(strippedframe) < len(myfile2)) && !found {
			if myfile2[f2pos] != strippedframe[0] {
				f2pos++
			} else {
				if Equal(myfile2[f2pos:f2pos+len(strippedframe)], strippedframe) {
					frameLenInSecondFile = int(binary.BigEndian.Uint16(myfile2[f2pos-2 : f2pos]))
					fmt.Println("frame length in 2nd file: ", frameLenInSecondFile)
					fmt.Printf("not frame space: %d - %d = %d\n", laststop, f2pos, f2pos-laststop)
					if len(strippedframe) != frameLenInSecondFile {
						fmt.Printf("frame length not match! \nbyte difference in 2nd file before frame: %x\nbyte difference in 2nd file after frame: %x\n", myfile2[laststop:f2pos], myfile2[f2pos+len(strippedframe):f2pos+frameLenInSecondFile])
					}
					//fmt.Printf("%x\n", myfile2[f2pos-2:f2pos])
					fmt.Println("found!")
					laststop = f2pos + len(strippedframe)
					found = true

					f2pos += len(strippedframe)
				} else {
					// fmt.Println(myfile2[f2pos : f2pos+len(strippedframe)])
					// fmt.Println("==================")
					// fmt.Println(strippedframe)
					f2pos++
				}
				//time.Sleep(10 * time.Second)
			}
		}
		//fmt.Println(myfile2[posstopfile2:])
		if !found {
			fmt.Println("not found :c !")
		}
		framenumber++

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
