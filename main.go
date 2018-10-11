package main

import (
	"github.com/gin-gonic/gin"
	"flag"
	"github.com/gorilla/websocket"
	"time"
	"net/http"
	"log"
	"fmt"
	"sync"
)

var (
	listen = flag.String("listen", ":12312", "server listen port(example :12312)")
	message = make(map[string]chan interface{})
	client = make(map[string]*websocket.Conn)
	mux sync.Mutex
)

func main() {
	r := gin.Default()
	// 连接ws会先发Get，正常返回101
	r.GET("/client/:id", func(c *gin.Context) {
		id := c.Param("id")
		//if _, exist := getClient(id); exist {
		//	c.JSON(http.StatusLocked, gin.H{
		//		"message": fmt.Sprintf("client %s is locked!", id),
		//	})
		//	return
		//}
		WsHandler(c.Writer, c.Request, id)
	})

	r.DELETE("/client/:id", func(c *gin.Context) {
		id := c.Param("id")
		if conn, exist := getClient(id); exist {
			conn.Close()
			deleteClient(id)
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"message": fmt.Sprintf("client %s is not found!", id),
			})
		}
		if _, exist := getMessageChannel(id); exist {
			deleteMessageChannel(id)
		}
	})
	
	r.POST("/message/:id", messageHandle)
	r.POST("/message", messageHandle)
	
	r.Run(*listen)
}

func messageHandle(c *gin.Context) {
	id := c.Param("id")
	if id != "" {
		_, exist := getMessageChannel(id)
		if !exist {
			c.JSON(http.StatusNotFound, gin.H{
				"message": fmt.Sprintf("not exist this client %s", id),
			})
			return
		}
	}

	var m interface{}

	if err := c.BindJSON(&m); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "message set failed",
		})
		return
	}
	if id == "" {
		setMessageAllClient(m)
	} else {
		setMessage(id, m)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "send success",
	})
	return
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	HandshakeTimeout: 5 * time.Second,
	// 取消ws跨域校验
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 处理ws请求
func WsHandler(w http.ResponseWriter, r *http.Request, id string) {
	var conn *websocket.Conn
	var err error
	pingTicker := time.NewTicker(time.Second * 10)
	conn, err = wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to set websocket upgrade: %+v", err)
		return
	}
	addClient(id, conn)
	m, exist := getMessageChannel(id)
	if !exist {
		m = make(chan interface{})
		addMessageChannel(id, m)
	}

	conn.SetCloseHandler(func(code int, text string) error {
		deleteClient(id)
		fmt.Println(code)
		return nil
	})

	for {
		select {
		case content, ok := <- m:
			err = conn.WriteJSON(content)
			if err != nil {
				log.Println(err)
				if ok {
					go func() {
						m <- content
					}()
				}

				conn.Close()
				deleteClient(id)
				return
			}
		case <-pingTicker.C:
			conn.SetWriteDeadline(time.Now().Add(time.Second * 20))
			if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Println("send ping err:", err)
				conn.Close()
				deleteClient(id)
				return
			}
		}
	}

}

func addClient(id string, conn *websocket.Conn) {
	mux.Lock()
	client[id] = conn
	mux.Unlock()
}

func getClient(id string) (conn *websocket.Conn, exist bool) {
	mux.Lock()
	conn, exist = client[id]
	mux.Unlock()
	return
}

func deleteClient(id string) {
	mux.Lock()
	delete(client, id)
	mux.Unlock()
}

func addMessageChannel(id string, m chan interface{}) {
	mux.Lock()
	message[id] = m
	mux.Unlock()
}

func getMessageChannel(id string) (m chan interface{}, exist bool) {
	mux.Lock()
	m, exist = message[id]
	mux.Unlock()
	return
}

func setMessage(id string, content interface{}) {
	mux.Lock()
	if m, exist := message[id]; exist {
		go func() {
			m <- content
		}()
	}
	mux.Unlock()
}

func setMessageAllClient(content interface{})  {
	mux.Lock()
	all := message
	mux.Unlock()
	go func() {
		for _, m := range all {
			m <- content
		}
	}()

}

func deleteMessageChannel(id string) {
	mux.Lock()
	if m, ok := message[id]; ok {
		close(m)
		delete(message, id)
	}
	mux.Unlock()
}