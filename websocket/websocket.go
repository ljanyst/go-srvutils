//------------------------------------------------------------------------------
// Author: Lukasz Janyst <lukasz@jany.st>
// Date: 06.11.2019
//
// Licensed under the MIT License, see the LICENSE file for details.
//------------------------------------------------------------------------------

package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// equalASCIIFold returns true if s is equal to t with ASCII case folding as
// defined in RFC 4790 - this comes from "github.com/gorilla/websocket"
func equalASCIIFold(s, t string) bool {
	for s != "" && t != "" {
		sr, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		tr, size := utf8.DecodeRuneInString(t)
		t = t[size:]
		if sr == tr {
			continue
		}
		if 'A' <= sr && sr <= 'Z' {
			sr = sr + 'a' - 'A'
		}
		if 'A' <= tr && tr <= 'Z' {
			tr = tr + 'a' - 'A'
		}
		if sr != tr {
			return false
		}
	}
	return s == t
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header["Origin"]
		if len(origin) == 0 {
			return true
		}

		u, err := url.Parse(origin[0])
		if err != nil {
			return false
		}

		if strings.HasPrefix(u.Host, "localhost") {
			return true
		}

		return equalASCIIFold(u.Host, r.Host)
	},
}

type RequestHandler interface {
	NewClient() []Response
	ProcessRequest(req Request) []Response
}

type WebSocketHandler struct {
	ctrl       *controller
	requestMap map[string]reflect.Type
}

func readMessages(requestMap map[string]reflect.Type, conn *websocket.Conn, l *link) {
	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			l.close()
			return
		}

		if messageType != websocket.TextMessage {
			continue
		}

		var request Request
		request = &RequestHeader{}
		err = json.Unmarshal(data, request)
		if err != nil {
			log.Error("Unable to unmarshal request: ", err)
			continue
		}

		if request.Type() != ACTION {
			log.Error("Malformed request: Not an action")
			continue
		}

		if request.Id() == "" {
			log.Error("Malformed request: No ID")
			continue
		}

		if t, ok := requestMap[request.Action()]; ok {
			obj := reflect.New(t.Elem()).Interface()
			request = obj.(Request)
		} else {
			request = &GenericRequest{}
		}

		err = json.Unmarshal(data, request)
		if err != nil {
			log.Error("Unable to unmarshal request: ", err)
			continue
		}

		select {
		case l.requestChan <- request:
		case <-l.closeChan:
			return
		}
	}

}

func writeMessages(conn *websocket.Conn, l *link) {
	for {
		select {
		case resp := <-l.responseChan:
			data, err := json.Marshal(resp)
			if err != nil {
				log.Error("Unable to marshal response: ", err)
				continue
			}
			err = conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				l.close()
				return
			}
		case <-l.closeChan:
			return
		}
	}
}

func (handler WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Unable to upgrade: ", err)
		return
	}

	l := handler.ctrl.getLink()

	go readMessages(handler.requestMap, conn, l)
	writeMessages(conn, l)
}

func (handler WebSocketHandler) InjectRequest(req Request) {
	handler.ctrl.injectRequest(req)
}

func NewWebSocketHandler(
	handler RequestHandler,
	requestMap map[string]reflect.Type) (WebSocketHandler, error) {

	var wsHandler WebSocketHandler

	reqType := reflect.TypeOf((*Request)(nil)).Elem()
	for n, t := range requestMap {
		if !t.Implements(reqType) {
			return wsHandler,
				fmt.Errorf(
					"Type %s for action %s does not implement the Request interface",
					t.Name(), n)
		}
	}

	wsHandler.ctrl = newController(handler)
	wsHandler.requestMap = requestMap
	return wsHandler, nil
}
