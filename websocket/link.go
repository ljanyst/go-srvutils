//------------------------------------------------------------------------------
// Author: Lukasz Janyst <lukasz@jany.st>
// Date: 06.11.2019
//
// Licensed under the MIT License, see the LICENSE file for details.
//------------------------------------------------------------------------------

package websocket

type link struct {
	requestChan  chan Request
	responseChan chan Response
	closeChan    chan bool
}

func (l *link) close() {
	// At any time we may have two goroutines that may wait for the channel
	// closure
	l.closeChan <- true
	l.closeChan <- true
}

func newLink() *link {
	l := new(link)
	l.requestChan = make(chan Request, 25)
	l.responseChan = make(chan Response, 25)
	l.closeChan = make(chan bool, 2)
	return l
}
