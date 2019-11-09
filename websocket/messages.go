//------------------------------------------------------------------------------
// Author: Lukasz Janyst <lukasz@jany.st>
// Date: 06.11.2019
//
// Licensed under the MIT License, see the LICENSE file for details.
//------------------------------------------------------------------------------

package websocket

type Request interface {
	Id() string
	Type() RequestType
	Action() string
}

type RequestType string

const (
	ACTION RequestType = "ACTION"
)

type RequestHeader struct {
	ReqId     string      `json:"id"`
	ReqType   RequestType `json:"type"`
	ReqAction string      `json:"action"`
}

func (r *RequestHeader) Id() string {
	return r.ReqId
}

func (r *RequestHeader) Type() RequestType {
	return r.ReqType
}

func (r *RequestHeader) Action() string {
	return r.ReqAction
}

type GenericRequest struct {
	RequestHeader
	Payload interface{}
}

type ResponseType string

const (
	UNICAST   ResponseType = "UNICAST"
	BROADCAST              = "BROADCAST"
	STATUS                 = "STATUS"
)

type ResponseStatus string

const (
	OK    ResponseStatus = "OK"
	ERROR                = "ERROR"
)

type Response struct {
	Type    ResponseType   `json:"type"`
	Status  ResponseStatus `json:"status"`
	Payload interface{}    `json:"payload"`
	Id      string         `json:"id"`
}
