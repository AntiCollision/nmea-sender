package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/adrianmo/go-nmea"
	"github.com/tarm/serial"
)

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
	defer s.Close()

	for {
		if read.Scan() {
			text := read.Text()
			nm, err := nmea.Parse(text)
			if err != nil {
				log.Println(err)
			} else {
				log.Println(nm)
			}
		}
	}
	_, err = net.Dial("tcp", fmt.Sprintf("%s:%d", *sendIP, *sendPort))

	if err != nil {
		log.Fatal(err)
	}
}
