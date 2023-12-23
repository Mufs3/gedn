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

func readFromConn(conn net.Conn) string {
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Cant read from conn by error: %s", err)
		conn.Close()
	}
	return string(buffer)
}

func init_client() net.Listener {
	listener, err := net.Listen("tcp", ":"+Port)
	if err != nil {
		listener.Close()
		log.Fatalf("Unusable port %s by error: %s \n", Port, err)
	}
	return listener
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

func writeMessage(conn net.Conn, connected_users map[net.Addr]string, output chan TMessage) {
	for {
		str := readFromConn(conn)
		output <- TMessage{conn.RemoteAddr(), str}
	}
}

func readMessage(conn net.Conn, connected_users map[net.Addr]string, input chan TMessage) {
	for {
		message := <-input
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

func handle_conn(conn net.Conn, connected_users map[net.Addr]string, messages chan TMessage) {
	init_conn(conn, connected_users)

	user_name := connected_users[conn.RemoteAddr()]

	write2Conn(conn, "Hi "+user_name)

	go writeMessage(conn, connected_users, messages)
	go readMessage(conn, connected_users, messages)

	recover()
	return
}

func main() {
	listener := init_client()
	connected_users := make(map[net.Addr]string)
	messages := make(chan TMessage)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Can't accept connection in port %s by error: %s\n", Port, err)
		}
		go handle_conn(conn, connected_users, messages)
	}
}
