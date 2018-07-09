package main

import (
	"io"
	"net"
	"log"
	"fmt"
	"bufio"
)

type User struct {
	Name   string
	Output chan Message
}

type Message struct {
	Username string
	Text     string
}

type Chatik struct {
	Users map[string]User
	Join  chan User
	Leave chan User
	Input chan Message
}

func (cs *Chatik) Run() {
	for {
		select {
		case user := <-chat.join:
			chat.users.mux.Lock()
			chat.users.allUsers[user.nickname] = user
			go func() {
				cs.Input <- Message{
					Username: "Admin",
					Text:     fmt.Sprintf("%s joined Chatik ", user.Name),
				}
			}()
			chat.users.mux.Unlock()
		case user := <-cs.Leave:
			chat.users.mux.Lock()
			delete(chat.users.allUsers, user.nickname)
			go func() {
				cs.Input <- Message{
					Username: "Admin",
					Text:     fmt.Sprintf("%s left Chatik", user.Name),
				}
			}()
			chat.users.mux.Unlock()
		case msg := <-cs.Input:
			chat.users.mux.Lock()
			for _, user := range cs.Users {
				select {
				case user.Output <- msg:
				default:
				}
			}
			chat.users.mux.Unlock()
		}
	}
}

func handleConn(chatServer *Chatik, conn net.Conn) {
	defer conn.Close()

	io.WriteString(conn, "Enter your Username:")

	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	user := User{
		Name:   scanner.Text(),
		Output: make(chan Message, 10),
	}
	chatServer.Join <- user
	defer func() {
		chatServer.Leave <- user
	}()

	// Read from conn
	go func() {
		for scanner.Scan() {
			ln := scanner.Text()
			chatServer.Input <- Message{user.Name, ln}
		}
	}()

	// write to conn
	for msg := range user.Output {
		if msg.Username != user.Name {
			_, err := io.WriteString(conn, msg.Username+": "+msg.Text+"\n")
			if err != nil {
				break
			}
		}
	}
}

func main() {
	server, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer server.Close()

	chatServer := &Chatik{
		Users: make(map[string]User),
		Join:  make(chan User),
		Leave: make(chan User),
		Input: make(chan Message),
	}
	go chatServer.Run()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatalln(err.Error())
		}
		go handleConn(chatServer, conn)
	}
}
