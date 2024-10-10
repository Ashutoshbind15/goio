package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/cors"
	"golang.org/x/net/websocket"
)

type HttpRes struct {
	Code int 
	Msg string
}

type WSMsg struct {
	MsgType string
	MsgData string
	RoomId string
}

type Server struct {
	rooms map[string][]*websocket.Conn
	conns map[*websocket.Conn]*ConnectionData
}

type ConnectionData struct {
	room string
	uid uuid.UUID
	isconn bool
}

type RoomMessage struct {
	Uid uuid.UUID
	Text string
}

type ctxkey struct{}

func middleFunc(nxt http.HandlerFunc) http.HandlerFunc {
	hf := func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("running the mware")

		var uid = uuid.New()
		ctx := context.WithValue(r.Context(), ctxkey{}, uid)

        nxt.ServeHTTP(w, r.WithContext(ctx))
    }

	return hf
}

func wsmware(wsHandler websocket.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsHandler.ServeHTTP(w, r)
	}
}

func hellofunc(w http.ResponseWriter, rq *http.Request) {

	rs := HttpRes{
		Code: 200,
		Msg: "scs",
	}

	res,err := json.Marshal(rs)

	if err != nil {
		panic(err)
	}

	w.Write(res)
}

func NewServer() *Server {
	return &Server{
		rooms: make(map[string][]*websocket.Conn),
		conns: make(map[*websocket.Conn]*ConnectionData),
	}
}

func uidfromwsrq (ws *websocket.Conn) uuid.UUID {
	return ws.Request().Context().Value(ctxkey{}).(uuid.UUID)
}

func convertToMsgStream (uid uuid.UUID, txt string) []byte {
	msg := RoomMessage{
		Uid: uid,
		Text: txt,
	}

	msgstr, err := json.Marshal(msg)

	if err != nil {
		fmt.Println("Error in converting to msg stream: ", err)
		return make([]byte, 0)
	}

	return msgstr

}

func (s *Server) handleWs(ws *websocket.Conn) {
	// fmt.Println("new conn from client", ws.RemoteAddr())
	
	s.conns[ws] = &ConnectionData{
		room: "",
		uid: uidfromwsrq(ws),
		isconn: true,
	}

	// add a mutex

	defer func() {		
		// s.conns[ws].isconn = false
		delete(s.conns, ws)
		ws.Close()
	}()
	
	s.readLoop(ws)
}

func (s *Server) readLoop(ws *websocket.Conn) {
	buff := make([]byte, 1024)

	for {
		n,err := ws.Read(buff)

		if err != nil {

			if err == io.EOF {
				break
			}

			fmt.Println("read err: ", err)
			continue
		}

		msg := buff[:n]

		var data WSMsg
		err = json.Unmarshal(msg, &data)

		fmt.Println(data)

		if err != nil {
			panic(err)
		}

		switch {
			case data.MsgType == "JOIN":
				s.rooms[data.RoomId] = append(s.rooms[data.RoomId], ws)
				s.roomSend(data.RoomId, convertToMsgStream(uidfromwsrq(ws), data.MsgData))
			
			case data.MsgType == "GLOBALCHAT":
				s.broadCast(convertToMsgStream(uidfromwsrq(ws), data.MsgData))
			
			case data.MsgType == "ROOMCHAT":
				s.roomSend(data.RoomId, convertToMsgStream(uidfromwsrq(ws), data.MsgData))
		}

	}
}

func (s* Server) broadCast (b []byte) {
	for conn, val:= range s.conns {
		if val.isconn {
			go func(){
				if _, err := conn.Write(b); err != nil {
					fmt.Println("err in writing: ", err)
				}
			}()
		}
	}
}

func (s *Server) roomSend (roomid string, b []byte) {
	for _,conn := range s.rooms[roomid] {
		cdata := s.conns[conn]

		if cdata.isconn {
			go func(){
				if _, err := conn.Write(b); err != nil {
					fmt.Println("err in writing: ", err)
				}
			}()
		}
	} 
}


func main() {

	// godotenv.Load()
	// dbclient := data.ConnectDb()

	// defer func() {
	// 	if err := dbclient.Disconnect(context.TODO()); err != nil {
	// 		panic(err)
	// 	}
	// }()


	// demoDbCall(dbclient)

	c := cors.New(cors.Options{
        AllowedOrigins:   []string{"*"}, // You can specify allowed origins here
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Authorization", "Content-Type"},
        AllowCredentials: true,
    })


	loggedHello := middleFunc(hellofunc)
	http.Handle("/", c.Handler(loggedHello))
	
	svr := NewServer()
	
	http.Handle("/ws", c.Handler(middleFunc(wsmware(websocket.Handler(svr.handleWs)))))
	http.ListenAndServe(":3000", nil)
}