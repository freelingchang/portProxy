package main

import (
	"net"
	"log"
	"fmt"
	"os"
	"strings"
	"io"
	"time"
)

func main(){
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	service := ":33128"
	clientIP := "127.0.0.1"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)
	clients := make(chan net.Conn)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("建立连接错误")
			continue
		}
		go Handle(conn,clients,clientIP)
	}
}

func Handle(conn net.Conn,clients chan net.Conn,clientIP string){
	ip := strings.Split(conn.RemoteAddr().String(),":")
	if len(ip) < 2 {
		defer conn.Close()
		log.Println("客户端IP错误")
		return
	}
	if ip[0] == clientIP {
		log.Println("client 发起连接,建立隧道")
		clients <- conn
		return
	}
	one := make([]byte,1024)
	n,err := conn.Read(one)
	if err != nil {
		return
	}
	client ,buff := getConn(clients,one[0:n])
	_,err = conn.Write(buff)
	if err != nil {
		log.Println("建立连接失败")
	}

	log.Println("双向连接建立，开始转发",client.RemoteAddr())
	go fdCopy(client, conn)
	go fdCopy(conn, client)
}

func getConn(clients chan net.Conn,one []byte) (net.Conn,[]byte){
	for {
		client := <- clients
		client = getConnWrite(client, one)
		if client == nil {
			continue
		}
		buff := make([]byte, 1024)
		client, buf := getConnRead(client, buff)
		if client == nil {
			continue
		}else{
			return client, buf
		}
	}

}

func getConnRead(client net.Conn,one []byte) (net.Conn,[]byte) {
	client.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	n, err := client.Read(one)
	if err != nil {
		log.Println("连接已关闭")
		return nil,nil
	}
	if n != len(one) {
		log.Println("连接已关闭")
		return nil,nil
	} else {
		log.Println("测试连接成功")
		client.SetReadDeadline(time.Time{})
		return client,one[0:n]
	}
}

func getConnWrite(client  net.Conn,one []byte) (net.Conn) {
	client.SetWriteDeadline(time.Now().Add(100 * time.Millisecond))
	n, err := client.Write(one)
	if err != nil {
		log.Println("连接已关闭")
		return nil
	}
	if n != len(one) {
		log.Println("连接已关闭")
		return nil
	} else {
		log.Println("测试连接成功")
		client.SetWriteDeadline(time.Time{})
		return client
	}
}

func fdCopy(f net.Conn,t net.Conn)  {
	defer f.Close()
	defer t.Close()
	_,err := io.Copy(f,t)
	if err != nil {
		//log.Println(err)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}