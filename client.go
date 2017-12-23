package main

import (
    "net"
    "log"
    "time"
    "io"
)

func main() {
    log.SetFlags(log.Lshortfile | log.LstdFlags)
    service := "127.0.0.1:33128"
    localSrv := "172.16.0.101:80"
    clients := make(chan net.Conn, 20)
    clientNum := make(chan int, 20)
    createChan(5, clientNum)
    go clientHandle(clients, localSrv, clientNum)
    keepConnectNum(service, clientNum, clients)

}

func clientHandle(clients chan net.Conn, localSrv string, clientNum chan int) {
    for {
        client := <-clients
        go forwardLocal(client, localSrv, clientNum)
    }
}

func forwardLocal(client net.Conn, localSvc string, clientNum chan int) {
    defer insertNum(clientNum)
    buff := make([]byte, 4096)
    n, err := client.Read(buff)
    if err != nil {
        log.Println("第一次读取数据失败", err)
        time.Sleep(10 * time.Second)
        return
    }
    //log.Println(string(buff[0:n]))
    tcpAddr, err := net.ResolveTCPAddr("tcp4", localSvc)
    conn, err := net.DialTCP("tcp", nil, tcpAddr)
    if err != nil {
        time.Sleep(10 * time.Second)
        return
    }
    n, err = conn.Write(buff[0:n])
    if err != nil {
        log.Println("转发错误")
        return
    }
    go fdCopy(client, conn)
    go fdCopy(conn, client)
}

func insertNum(clientNum chan int) {
    clientNum <- 1
}

func createChan(n int, number chan int) {
    for i := 0; i < n; i++ {
        number <- 1
    }
    log.Println("初始化chan int")
}

func connect(clientIP string, clients chan net.Conn, clientNum chan int) {
    tcpAddr, err := net.ResolveTCPAddr("tcp4", clientIP)
    conn, err := net.DialTCP("tcp", nil, tcpAddr)
    if err != nil {
        time.Sleep(10 * time.Second)
        clientNum <- 1
        log.Println("连接失败", clientIP)
        return
    }
    clients <- conn
    log.Println("建立连接成功", conn.LocalAddr())
}

func keepConnectNum(clientIP string, num chan int, clients chan net.Conn) {
    for {
        select {
        case _ = <-num:
            go connect(clientIP, clients, num)
        }
    }
}

func fdCopy(f net.Conn, t net.Conn) {
    defer f.Close()
    defer t.Close()
    _, err := io.Copy(f, t)
    if err != nil {
        //log.Println(err)
    }
}

func getConn(clients chan net.Conn) net.Conn {
    for {
        select {
        case client := <-clients:
            client.SetReadDeadline(time.Now())
            one := []byte{}
            if _, err := client.Read(one); err == io.EOF {
                log.Println("连接已关闭")
            } else {
                client.SetReadDeadline(time.Time{})
                return client
            }
        }
    }
}

func checkConn(client net.Conn) bool {
    return true
    client.SetReadDeadline(time.Now())
    one := []byte{}
    if _, err := client.Read(one); err == io.EOF {
        log.Println("连接已关闭")
        return true
    } else {
        client.SetReadDeadline(time.Time{})
        return true
    }
}
