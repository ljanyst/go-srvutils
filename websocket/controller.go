//------------------------------------------------------------------------------
// Author: Lukasz Janyst <lukasz@jany.st>
// Date: 06.11.2019
//
// Licensed under the MIT License, see the LICENSE file for details.
//------------------------------------------------------------------------------

package websocket

import (
	"sync/atomic"

	log "github.com/sirupsen/logrus"
)

const (
	addClient     = 0
	removeClient  = 1
	injectRequest = 2
)

type requestWrapper struct {
	clientId uint64
	request  Request
}

type ctrl struct {
	action       uint
	id           uint64
	responseChan chan<- Response
	sync         chan bool
}

type controller struct {
	lastLinkId   uint64
	handler      RequestHandler
	broadcastMap map[uint64]chan<- Response
	requestChan  chan requestWrapper
	controlChan  chan ctrl
}

func (c *controller) getLink() *link {
	l := newLink()
	linkId := atomic.AddUint64(&c.lastLinkId, 1)
	sync := make(chan bool)
	c.controlChan <- ctrl{addClient, linkId, l.responseChan, sync}
	<-sync

	go func() {
		for {
			select {
			case <-l.closeChan:
				c.controlChan <- ctrl{removeClient, linkId, nil, sync}
				<-sync
				return
			case req := <-l.requestChan:
				c.requestChan <- requestWrapper{linkId, req}
			}
		}
	}()

	return l
}

func (c *controller) sendMessages(clientId uint64, reqId string, msgs []Response) {
	channel, ok := c.broadcastMap[clientId]

	for _, msg := range msgs {
		if msg.Type == BROADCAST {
			for _, channel := range c.broadcastMap {
				channel <- msg
			}
		} else {
			if ok {
				if msg.Type == STATUS {
					msg.Id = reqId
				}
				channel <- msg
			} else {
				log.Errorf("Trying to send a message to a non-existing client:", clientId)
			}
		}
	}
}

func (c *controller) handleRequests() {
	for {
		select {
		case req := <-c.requestChan:
			msgs := c.handler.ProcessRequest(req.request)
			c.sendMessages(req.clientId, req.request.Id(), msgs)
		case ctrl := <-c.controlChan:
			switch ctrl.action {
			case addClient:
				c.broadcastMap[ctrl.id] = ctrl.responseChan
				ctrl.sync <- true
				msgs := c.handler.NewClient()
				c.sendMessages(0, "", msgs)
			case removeClient:
				delete(c.broadcastMap, ctrl.id)
				ctrl.sync <- true
			}
		}
	}
}

func (c controller) injectRequest(req Request) {
	c.requestChan <- requestWrapper{0, req}
}

func newController(handler RequestHandler) *controller {
	c := new(controller)
	c.lastLinkId = 0
	c.handler = handler
	c.broadcastMap = make(map[uint64]chan<- Response, 250)
	c.requestChan = make(chan requestWrapper, 100)
	c.controlChan = make(chan ctrl)
	go c.handleRequests()
	return c
}
