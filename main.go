package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	nmea "github.com/adrianmo/go-nmea"
	"github.com/tarm/serial"
)

func sender(conn net.Conn, data chan nmea.Sentence) {
	pos := Positioning{}
	mu := sync.Mutex{}

	go func() {
		data := make([]byte, 28)
		startTime := time.Now().Unix()
		for {
			reader := bytes.NewBuffer(data)
			durTime := time.Now().Unix() - startTime

			mu.Lock()
			pos.Time = int32(durTime)
			mu.Unlock()
			err := binary.Write(reader, binary.BigEndian, pos)
			if err != nil {
				log.Println("ERROR! : ", err)
			}

			log.Println(data)
			time.Sleep(time.Second * 1)
		}
	}()

	for {
		recv := <-data
		log.Println(recv.String())

		mu.Lock()
		switch recv.DataType() {
		case nmea.TypeXDR:
			xdr := recv.(nmea.XDR)
			log.Println(xdr)
			log.Println("THIS XDR")
			pos.Roll = float32(xdr.Measurements[0].Value)
			pos.Pitch = float32(xdr.Measurements[1].Value)
			break
		case nmea.TypeGLL:
			gll := recv.(nmea.GPGLL)
			log.Println(gll)
			pos.X = float32(gll.Latitude)
			pos.Y = float32(gll.Longitude)
			break
		case nmea.TypeROT:
			rot := recv.(nmea.ROT)
			log.Println(rot)
			pos.Yaw = float32(rot.RateOfTurn)
			break
		}
		mu.Unlock()
	}
}

func main() {
	comPort := flag.String("c", "", "COM 포트 정보를 입력해 주세요.")
	baudRate := flag.Int("b", 115200, "Baudrate 정보를 입력해 주세요.")
	sendIP := flag.String("i", "", "전송할 IP 정보를 입력해 주세요.")
	sendPort := flag.Int("p", 0, "전송할 Port 정보를 입력해 주세요.")
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

	server, err := net.Dial("udp", fmt.Sprintf("%s:%d", *sendIP, *sendPort))
	if err != nil {
		log.Fatal(err)
	}
	reqChan := make(chan nmea.Sentence)
	go sender(server, reqChan)

	defer s.Close()

	for {
		if read.Scan() {
			text := read.Text()
			nm, err := nmea.Parse(text)
			if err != nil {
				// log.Println(err)
			} else {
				// log.Println(nm)
				reqChan <- nm
			}
		}
	}
}
