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

func main() {
    log.SetFlags(log.Lshortfile | log.LstdFlags)
    service := ":33128"
    controlBind := ":33129"
    clientIP := "192.168.1.3"
    tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
    checkError(err)
    listener, err := net.ListenTCP("tcp", tcpAddr)
    checkError(err)
    clients := make(chan net.Conn,20)
    controlChan := make(chan int,20)
    go controlThread(controlBind,controlChan)
    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Println("建立连接错误")
            continue
        }
        go Handle(conn, clients, clientIP,controlChan)
    }
}

func controlThread(bind string,n chan int)  {
    tcpAddr, err := net.ResolveTCPAddr("tcp4", bind)
    checkError(err)
    listener, err := net.ListenTCP("tcp", tcpAddr)
    checkError(err)
    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Println("建立连接错误")
            continue
        }
        log.Println("控制通道连接进入")
        go controlHandle(conn, n)
    }
}
func controlHandle(conn net.Conn,n chan int){
    defer conn.Close()
    for {
        select {
            case _ = <- n:
                conn.SetWriteDeadline(time.Now().Add(time.Second * 2))
                _, err := conn.Write([]byte("1\r\n"))
                if err != nil {
                    log.Println("控制消息发送失败")
                    n <- 1
                    return
                }
                conn.SetReadDeadline(time.Now().Add(time.Second * 2))
                rep := make([]byte,64)
                _, err = conn.Read(rep)
                if err != nil {
                    log.Println("控制消息发送收取回复失败")
                    n <- 1
                    return
                }
                log.Println("发送控制消息1次")
        }
    }

}

func Handle(conn net.Conn, clients chan net.Conn, clientIP string,controlChan chan int) {
    ip := strings.Split(conn.RemoteAddr().String(), ":")
    if len(ip) < 2 {
        defer conn.Close()
        log.Println("客户端IP错误")
        return
    }
    if strings.HasPrefix(ip[0],"100") {
        defer conn.Close()
        log.Println("阿里云服务器,关闭连接")
        return
    }
    if ip[0] == clientIP {
        log.Println("client 发起连接,建立隧道")
        clients <- conn
        return
    } else {
        log.Println("一般客户端连接,开始转发", ip)
        controlChan <- 1
        client := <-clients
        go copyDefer(client, conn)
    }
}

func copyDefer(f net.Conn, t net.Conn) {
    defer f.Close()
    defer t.Close()
    c := make(chan int,2)
    go fdCopy(f, t,c)
    go fdCopy(t, f,c)
    _ = <- c

}


func fdCopy(f net.Conn, t net.Conn,c chan int) {
    defer f.Close()
    defer t.Close()
    _, err := io.Copy(f, t)
    if err != nil {
        c <- 1
    }
}

func checkError(err error) {
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
        os.Exit(1)
    }
}
