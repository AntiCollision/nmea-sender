package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/tarm/serial"
)

func main() {
	comPort := flag.String("c", "", "COM 포트 정보를 입력해 주세요.")
	baudRate := flag.Int("b", 115200, "Baudrate 정보를 입력해 주세요.")
	sendIP := flag.String("i", "", "전송할 IP 정보를 입력해 주세요.")
	sendPort := flag.Int("p", 0, "전송할 Port 정보를 입력해 주세요.")

	fmt.Println("ComPort : ", *comPort)
	fmt.Println("Baudrate : ", *baudRate)
	fmt.Println("Send IP : ", *sendIP)
	fmt.Println("Send Port : ", *sendPort)

	c := &serial.Config{Name: *comPort, Baud: *baudRate}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	buf := make([]byte, 1024)

	n, err := s.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%q", buf[:n])

	_, err = net.Dial("tcp", fmt.Sprintf("%s:%d", *sendIP, *sendPort))

	if err != nil {
		log.Fatal(err)
	}
}
