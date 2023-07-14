package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"sync"
	"time"
	"unsafe"

	nmea "github.com/adrianmo/go-nmea"
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
			// log.Println(xdr)
			// log.Println("THIS XDR")
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

	// c := &serial.Config{Name: *comPort, Baud: *baudRate}
	// s, err := serial.OpenPort(c)
	// read := bufio.NewScanner(s)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	server, err := net.Dial("udp", fmt.Sprintf("%s:%d", *sendIP, *sendPort))
	if err != nil {
		log.Fatal(err)
	}
	reqChan := make(chan nmea.Sentence)
	go sender(server, reqChan)

	// defer s.Close()

	for {
		xdr, err := nmea.Parse("$IIXDR,A,0.7,D,ROLL,A,-1.9,D,PITCH*37")
		if err != nil {
			// log.Println(err)
		} else {
			// log.Println(nm)
			reqChan <- xdr
		}
		gpgll, err := nmea.Parse("$GPGLL,3506.2155,N,12904.9829,E,144347.00,A*02")
		if err != nil {
			// log.Println(err)
		} else {
			// log.Println(nm)
			reqChan <- gpgll
		}
		rot, err := nmea.Parse("$--ROT,-12.6,A*3E")
		if err != nil {
			log.Println(err)
		} else {
			reqChan <- rot
		}
		// if read.Scan() {
		// 	text := read.Text()
		// 	nm, err := nmea.Parse(text)

		// }
	}
}
