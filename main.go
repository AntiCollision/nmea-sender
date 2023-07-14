package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"sync"
	"time"
	"unsafe"

	nmea "github.com/adrianmo/go-nmea"
	"github.com/tarm/serial"
)

func sender(conn net.Conn, data chan nmea.Sentence) {
	pos := Positioning{}
	mu := sync.Mutex{}

	go func() {
		startTime := time.Now().Unix()
		for {
			durTime := time.Now().Unix() - startTime
			mu.Lock()
			pos.Time = int32(durTime)
			mu.Unlock()

			ptr := unsafe.Pointer(&pos)
			copy := *(*[28]byte)(ptr)
			dest := [28]byte{}
			for item := 0; item < 7; item += 1 {
				for mem := 0; mem < 4; mem += 1 {
					dest[item*4+mem] = copy[item*4+(3-mem)]
				}
			}

			size, err := conn.Write(dest[:])
			log.Println(dest, size, err)
			time.Sleep(time.Second * 1)
		}
	}()

	for {
		recv := <-data

		mu.Lock()
		switch recv.DataType() {
		case nmea.TypeXDR:
			xdr := recv.(nmea.XDR)
			log.Println(xdr)
			log.Println("THIS XDR")
			pos.Roll = float32(xdr.Measurements[0].Value)
			pos.Pitch = float32(xdr.Measurements[1].Value)
		case nmea.TypeGLL:
			gll := recv.(nmea.GLL)
			// log.Println(gll)
			pos.X = float32(gll.Latitude)
			pos.Y = float32(gll.Longitude)
		case nmea.TypeROT:
			rot := recv.(nmea.ROT)
			// log.Println(rot)
			pos.Yaw = float32(rot.RateOfTurn / 180 * math.Pi)
		}
		mu.Unlock()
	}
}

func main() {
	comPort := flag.String("c", "", "COM 포트 정보를 입력해 주세요.")
	baudRate := flag.Int("b", 115200, "Baudrate 정보를 입력해 주세요.")
	sendIP := flag.String("i", "127.0.0.1", "전송할 IP 정보를 입력해 주세요.")
	sendPort := flag.Int("p", 6000, "전송할 Port 정보를 입력해 주세요.")
	flag.Parse()

	fmt.Println("ComPort : ", *comPort)
	fmt.Println("Baudrate : ", *baudRate)
	fmt.Println("Send IP : ", *sendIP)
	fmt.Println("Send Port : ", *sendPort)

	c := &serial.Config{Name: *comPort, Baud: *baudRate}
	s, err := serial.OpenPort(c)
	read := bufio.NewScanner(s)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	server, err := net.Dial("udp", fmt.Sprintf("%s:%d", *sendIP, *sendPort))
	if err != nil {
		log.Fatal(err)
	}
	reqChan := make(chan nmea.Sentence)
	go sender(server, reqChan)

	for {
		if read.Scan() {
			text := read.Text()
			nm, err := nmea.Parse(text)
			if err != nil {
				log.Println(err)
			} else {
				reqChan <- nm
			}
		}
	}
}
