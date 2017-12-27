package main

import (
    "net"
    "log"
    "time"
    "io"
    "strings"
    "fmt"
    "os"
)

func main() {
    log.SetFlags(log.Lshortfile | log.LstdFlags)
    service := "192.168.1.230:33128"
    controlService := "192.168.1.230:33129"
    localSrv := "192.168.1.230:22"
    //clients := make(chan net.Conn, 100)
    controlAlive := make(chan int ,1)
    clientNum := make(chan int, 10)
    go clientHandle(service, localSrv, clientNum)
    //go ConnectForwardServer(service,,clients)
    controlAlive <- 1
    keepControlAlive(controlService, controlAlive,clientNum)

}

func clientHandle(service, localSrv string, clientNum chan int) {
    for {
        select {
        case _ = <-clientNum:
            log.Println("开始建立转发隧道")
            go forwardLocal(service, localSrv, clientNum)
        }
    }
}

func ConnectForwardServer(service string,clientNum chan int,clients chan net.Conn){
    for {
        select {
        case _ = <-clientNum:
            tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
            if err != nil {
                fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
                os.Exit(1)
            }
            conn, err := net.DialTCP("tcp", nil, tcpAddr)
            if err != nil {
                log.Println("连接服务器失败",service)
                time.Sleep(10 * time.Second)
                clientNum <- 1
                return
            }
            clients <- conn
        }
    }

}


func forwardLocal(service string,localSvc string, clientNum chan int) {
    tcpAddr, err := net.ResolveTCPAddr("tcp4", localSvc)
    conn, err := net.DialTCP("tcp", nil, tcpAddr)
    if err != nil {
        time.Sleep(10 * time.Second)
        log.Println("连接本地服务器失败")
        clientNum <- 1
        defer conn.Close()
        return
    }
    log.Println("连接本地服务器成功")
    serviceTcpAddr, err := net.ResolveTCPAddr("tcp4", service)
    client, err := net.DialTCP("tcp", nil, serviceTcpAddr)
    if err != nil {
        time.Sleep(10 * time.Second)
        log.Println("连接远程服务器失败")
        clientNum <- 1
        defer client.Close()
        return
    }
    log.Println("连接远程服务器成功")
    log.Println("开始转发")
    copyDefer(conn,client)
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

func recvControl(controlServer string, alive chan int, clientNum chan int) {
    defer insertAlive(alive)
    tcpAddr, err := net.ResolveTCPAddr("tcp4", controlServer)
    conn, err := net.DialTCP("tcp", nil, tcpAddr)
    if err != nil {
        time.Sleep(10 * time.Second)
        log.Println("连接失败", controlServer)
        return
    }
    defer conn.Close()
    log.Println("连接控制端口")
    for {
        buff := make([]byte,64)
        n,err := conn.Read(buff)
        if err != nil {
        	log.Println("控制通道失效，关闭连接")
        	return
        }
        log.Println("收到控制指令,长度为:",n)
        content := string(buff[:n])
        conLen := strings.Count(content,"\r\n")
        conn.SetWriteDeadline(time.Now().Add(time.Second * 2))
        _,err = conn.Write([]byte("ok\r\n"))
        if err != nil {
            log.Println("回复控制指令失败")
            return
        }
        if conLen == 0 && n >0 {
            log.Println("消息格式错误")
            return
        }
        for i :=0;i<conLen;i++ {
            log.Println("对端要求建立链接")
            clientNum <- 1
            log.Println("列队请求数加1")
        }

    }
}

func insertAlive(alive chan int){
	log.Println("重连")
    alive <- 1
}

func keepControlAlive(controlServer string, alive chan int, clientNum chan int) {
    for {
        select {
        case _ = <- alive:
            go recvControl(controlServer, alive,clientNum)
        }
    }
}

func copyDefer(f net.Conn, t net.Conn) {
    go fdCopy(f, t)
    go fdCopy(t, f)
}


func fdCopy(f net.Conn, t net.Conn) {
    defer f.Close()
    defer t.Close()
    _, err := io.Copy(f, t)
    if err != nil {
    }
}

