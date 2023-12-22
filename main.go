package main

import (
	"log"
	"net"
)

const Port = "8080"

func init_client() net.Listener {
	listener, err := net.Listen("tcp", ":"+Port)
	if err != nil {
		log.Fatalf("Unusable port %s by error: %s \n", Port, err)
	}
	return listener
}

func init_conn(conn net.Conn, connected_users map[net.Addr]string) {
	var buffer [1024]byte

	mess := []byte("Welcome, please provide a Name:\t")
	_, err := conn.Write(mess)
	if err != nil {
		log.Printf("Error on writing to conn: %s\n", err)
	}
	_, err = conn.Read(buffer[:])
	if err != nil {
		log.Printf("Error on writing to conn: %s\n", err)
	}
	connected_users[conn.RemoteAddr()] = string(buffer[:])
}

func handle_conn(conn net.Conn, connected_users map[net.Addr]string) {
	init_conn(conn, connected_users)
	user_name := connected_users[conn.RemoteAddr()]
	conn.Write([]byte("Hi " + user_name + "gt fckd\n"))
	conn.Close()
}

func main() {
	listener := init_client()
	connected_users := make(map[net.Addr]string)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Can't accept connection in port %s by error: %s\n", Port, err)
		}
		go handle_conn(conn, connected_users)
	}
}
