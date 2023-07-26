package main

import (
	"log"
	"os"

	//"syscall"
	"time"
	//"unsafe"
	"encoding/binary"
	"fmt"

	"golang.org/x/sys/unix"
)

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var msg1, msg2 unix.Msghdr
var previousarray, mypart, strippedframe []byte

/*	анализ хидера кадров:
	[0:4] 	+ не изменяется ffffffff, замена этого куска на случайное значение кажется ничего не меняет в выводе ffmpeg, файл проигрывается так же.
	[4:7]   + изменяется, возможно захватывает следующий байт, проверить, начинает не с 0 или 1, значение инкрементируется со временем даже если поток не тянется
	[7:8] 	+ не изменяется 00
	[8:10]	+ изменяется, проверить, значения не инкрементальны. Вероятно это размер пакета в литтл эндиан.
	[10:13] + не изменяется 000000
	[13:14]	+ изменяется, проверить, кажется значения могут быть только 10, 20, 40 (это hex)
	[14:68] + не изменяется 000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000
	[68:71] + изменяется, возможно захватывает следующий байт, проверить, начинает не с 0 или 1, значение инкрементируется со временем даже если поток не тянется
	[71:72] + не изменяется 00
	[72:77] + изменяется, проверить, начинает не с 0 или 1, значение инкрементируется со временем даже если поток не тянется
	[77:80] + не изменяется 7823e1, замена этого куска на случайное значение кажется ничего не меняет в выводе ffmpeg, файл проигрывается так же.
	[80:82] + изменяется, проверить, значения не инкрементальны
	[82:84] + не изменяется 0000
	[84:85] + изменяется, проверить, кажется значения могут быть только 00 и 01
	[85:100] + не изменяется 000000000000000000000000000001
	проверить - значит проверить, растут ли значения в кадрах инкрементально, с постоянными интервалами, не связаны ли с размером полезного пейлоада, насколько разнообразны значения


*/

func main() {
	mode := "replace_some_bytes" // "search_same_parts"/"show_parts"
	sleepBetweenFrames_processing := 1

	myfile1, _ := os.ReadFile("/home/reg/Desktop/развлечения/разборки с камерой/полученный с сокета поток/compared_part_my_own_binary_2.mp4")
	//myfile1, _ := os.ReadFile("/home/reg/Desktop/развлечения/разборки с камерой/полученный с сокета поток/socketvideo_another_2.mp4")

	fileWithReplacements, _ := os.Create("/tmp/fileWithReplacements.mp4")

	header := make([]byte, 100, 100)
	j := 0
	hasdifference := false

	binarylength := make([]byte, 4, 4)
	for i := 0; i < len(myfile1); i += 4096 {
		header = myfile1[i : i+100]
		currentframe := myfile1[i+100 : i+4096]
		mypartStart := 0
		mypartStop := 100
		mypart = header[mypartStart:mypartStop]
		newpart := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x11}
		fmt.Printf("\nframe number counter: %d\n", j)
		fmt.Printf("%x\n", header)

		for j := len(currentframe); j != 0; j-- {
			if currentframe[j-1] != 0 {
				strippedframe = currentframe[:j]
				binary.BigEndian.PutUint32(binarylength, uint32(len(strippedframe)))
				fmt.Printf("frame length: int %d, hex %x\n", len(strippedframe), binarylength)
				break
			}
		}

		switch mode {
		case "search_same_parts":
			if !comparebytemassive(mypart, &previousarray, &j) {
				fmt.Printf("found difference: %x - %x\n", mypart, previousarray)
				hasdifference = true
				break
			}
			previousarray = mypart
		case "show_parts":
			fmt.Printf("changed: %x %x %x %x %x %x %x\n", header[4:7], header[8:10], header[13:14], header[68:71], header[72:77], header[80:82], header[84:85])
		case "replace_some_bytes":
			fmt.Printf("trying to replace %x on %x\n", mypart, newpart)
			if len(mypart) != len(newpart) {
				fmt.Println("new part and old part size mismatch!")
				os.Exit(127)
			} else {
				fileWithReplacements.Write(myfile1[i : i+mypartStart])
				fileWithReplacements.Write(newpart)
				fileWithReplacements.Write(myfile1[i+mypartStop : i+4096])
			}

		}

		time.Sleep(time.Duration(sleepBetweenFrames_processing) * time.Millisecond)
		j++
	}
	if !hasdifference {
		fmt.Printf("no difference part: %x\n", mypart)
	}

}

func comparebytemassive(arraycurrent []byte, previousarray *[]byte, framenumber *int) bool {
	if *framenumber == 0 {
		return true
	} else {
		if Equal(arraycurrent, *previousarray) {
			return true

		} else {
			return false
		}

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
