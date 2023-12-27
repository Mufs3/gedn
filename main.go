package main

import (
	"log"
	"net"
	"strings"
)

const Port string = "8080"

type TMessage struct {
	Sender  net.Addr
	Message string
}

type Broadcaster struct {
	Input          chan TMessage
	OutputChannels *[]chan TMessage
}

func main() {
	listener := init_client()
	defer listener.Close()
	connected_users := make(map[net.Addr]string)
	var output_channels []chan TMessage = make([]chan TMessage, 1023)
	pChannels := &output_channels
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Can't accept connection in port %s by error: %s\n", Port, err)
		}
		channel := make(chan TMessage, 255)
		*pChannels = append(*pChannels, channel)
		broadcaster := Broadcaster{channel, pChannels}

		go handle_conn(conn, connected_users, broadcaster)
	}
}

func init_client() net.Listener {
	listener, err := net.Listen("tcp", ":"+Port)
	if err != nil {
		listener.Close()
		log.Fatalf("Unusable port %s by error: %s \n", Port, err)
	}
	return listener
}

func handle_conn(conn net.Conn, connected_users map[net.Addr]string, broadcaster Broadcaster) {
	init_conn(conn, connected_users)

	user_name := connected_users[conn.RemoteAddr()]

	write2Conn(conn, "Hi "+user_name+"\n")

	go broadcaster.WriteMessages(conn, connected_users)
	go broadcaster.ReadMessages(conn, connected_users)

	defer recover()

	return
}

func init_conn(conn net.Conn, connected_user map[net.Addr]string) error {
	buffer := make([]byte, 1023)
	write2Conn(conn, "Welcome, please provide a Name:\t")

	_, err := conn.Read(buffer[:])
	if err != nil {
		log.Printf("Error on reading conn: %s\n", err)
		return err
	}
	connected_user[conn.RemoteAddr()] = strings.Split(string(buffer), "\n")[0]

	log.Printf("connected to %s", conn.RemoteAddr().String())

	return nil
}

func (broadcaster Broadcaster) WriteMessages(conn net.Conn, connected_users map[net.Addr]string) {
	for {
		str := readFromConn(conn)

		for _, v := range *broadcaster.OutputChannels {
			if v != nil {
				v <- TMessage{conn.RemoteAddr(), str}
			}
		}
	}
}

func (broadcaster Broadcaster) ReadMessages(conn net.Conn, connected_users map[net.Addr]string) {
	for {
		message := <-broadcaster.Input
		log.Printf("recieved from %s:\t %s", message.Sender.String(), message.Message)
		if message.Sender != conn.RemoteAddr() {
			write2Conn(conn, formatMessage(message, connected_users))
		}
	}
}

func formatMessage(message TMessage, connected_users map[net.Addr]string) string {
	user_name := connected_users[message.Sender]
	str := user_name + ">>\t" + message.Message
	return str
}

func write2Conn(conn net.Conn, str string) {
	mess := []byte(str)
	_, err := conn.Write(mess)
	if err != nil {
		log.Printf("Error to writing to conn: %s", err)
	}
}

func readFromConn(conn net.Conn) string {
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		defer conn.Close()
		log.Panicf("Cant read from conn by error: %s\n Closing connection to %s", err, conn.RemoteAddr().String())
	}
	return string(buffer)
}
