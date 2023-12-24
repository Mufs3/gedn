package main

import (
	"log"
	"net"
)

const Port = "8080"

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
	connected_users := make(map[net.Addr]string)
	broadcasters := make([]Broadcaster, 1023)
	output_channels := make([]chan TMessage, 1023)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Can't accept connection in port %s by error: %s\n", Port, err)
		}
		broadcaster := Broadcaster{make(chan TMessage), &output_channels}
		broadcasters = append(broadcasters, broadcaster)

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

	write2Conn(conn, "Hi "+user_name)

	go broadcaster.WriteMessage(conn, connected_users)
	go broadcaster.ReadMessage(conn, connected_users)

	recover()
	return
}

func init_conn(conn net.Conn, connected_user map[net.Addr]string) {
	var buffer [64]byte

	write2Conn(conn, "Welcome, please provide a Name:\t")

	_, err := conn.Read(buffer[:])
	if err != nil {
		log.Printf("Error on reading conn: %s\n", err)
	}
	connected_user[conn.RemoteAddr()] = string(buffer[:])
}

func (broadcaster Broadcaster) WriteMessage(conn net.Conn, connected_users map[net.Addr]string) {
	for {
		err, str := readFromConn(conn)
		if err != nil {
			return
		}
		for _, v := range *broadcaster.OutputChannels {
			v <- TMessage{conn.RemoteAddr(), str}
		}
	}
}

func (broadcaster Broadcaster) ReadMessage(conn net.Conn, connected_users map[net.Addr]string) {
	for {
		message := <-broadcaster.Input
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
	n, err := conn.Write(mess)
	if err != nil {
		log.Printf("Error to writing to conn: %s", err)
	}
	for n < len(mess) {
		auxn, err := conn.Write(mess[n:])
		n = auxn + n
		if err != nil {
			log.Printf("Error to writing to conn: %s", err)
		}
	}
}

func readFromConn(conn net.Conn) (error, string) {
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Cant read from conn by error: %s", err)
		conn.Close()
	}
	return err, string(buffer)
}
