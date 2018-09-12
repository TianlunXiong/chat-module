package main

import (  

    "fmt"   

    "golang.org/x/net/websocket" 

    "html/template"              //支持模板html

    "log"

    "encoding/json"

    "net/http"  

    "github.com/garyburd/redigo/redis"

)

var users map[*websocket.Conn] string


func main() {

    users = make(map[*websocket.Conn]string)
    //接受websocket的路由地址

    http.Handle("/websocket", websocket.Handler(Echo))

    //html页面

    http.HandleFunc("/web", web)

    if err := http.ListenAndServe(":8080", nil); err != nil {

        log.Fatal("ListenAndServe:", err)

    }

}



func Echo(ws *websocket.Conn) {  
    var err error  

    c, err := redis.Dial("tcp", "127.0.0.1:6379")
    if err != nil {
        fmt.Println("Connect to redis error", err)
        return
    }
    defer c.Close()

    for {

        var reply string
        bag := make(map[int]string)
        
        if _, ok := users[ws]; !ok {
            users[ws] = "匿名用户"
        }

        //websocket接受信息  

        if err = websocket.Message.Receive(ws, &reply); err != nil {  

            fmt.Println("receive failed:", err)  
            // 收到EOF信号， 客户端 closed 从用户池清除这个连接
            delete(users, ws)
            break  

        }

        fmt.Println("reveived from client: " + reply)

        //写入Redis
        _, err = c.Do("LPUSH","20180912",reply)

        if err != nil {
            fmt.Println("redis set failed:", err)

            break
        }

        values, _ := redis.Values(c.Do("lrange", "20180912", "0", "100"))
        for i, v := range values {
            bag[i] = string(v.([]byte))
        }

        j, err := json.Marshal(bag)

        if err != nil {
            fmt.Println("转json failed", err)

            break
        }


        //这里是发送消息
        for key, _ := range users {
            if err = websocket.Message.Send(key, string(j)); err != nil {

                fmt.Println("send failed:", err)
                delete(users, key)
                break
    
            }
        }
    }

}

func web(w http.ResponseWriter, r *http.Request) {

    //打印请求的方法  

    fmt.Println("method", r.Method)

    if r.Method == "GET" { //如果请求方法为get显示login.html,并相应给前端  

        t, _ := template.ParseFiles("websocket.html")

        t.Execute(w, nil)  

    } else {  

        //否则走打印输出post接受的参数username和password  

        fmt.Println(r.PostFormValue("username"))  

        fmt.Println(r.PostFormValue("password"))  

    }  

}

