//------------------------------------------------------------------------------
// Author: Lukasz Janyst <lukasz@jany.st>
// Date: 07.09.2019
//
// Licensed under the MIT License, see the LICENSE file for details.
//------------------------------------------------------------------------------

package auth

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type attempt struct {
	timestamp   time.Time
	numAttemtps int
}

type BasicAuthHandler struct {
	userMap        map[string]string
	attemptMap     map[string]*attempt
	realm          string
	wrappedHandler http.Handler
	rnd            *rand.Rand
}

func (handler BasicAuthHandler) writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, handler.realm))
	w.WriteHeader(401)
	w.Write([]byte("Unauthorised.\n"))
}

func (handler BasicAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sleepTime := handler.rnd.Intn(500000)

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		time.Sleep(time.Duration(sleepTime) * time.Microsecond)
		handler.writeUnauthorized(w)
		return
	}

	att, ok := handler.attemptMap[ip]
	if !ok {
		att = new(attempt)
		handler.attemptMap[ip] = att
	}

	if att.numAttemtps >= 3 && att.timestamp.Add(5*time.Minute).After(time.Now()) {
		att.numAttemtps = 0
		att.timestamp = time.Now()
		time.Sleep(time.Duration(sleepTime) * time.Microsecond)
		handler.writeUnauthorized(w)
		return
	}

	user, pass, ok := r.BasicAuth()
	if !ok {
		att.numAttemtps++
		att.timestamp = time.Now()
		time.Sleep(time.Duration(sleepTime) * time.Microsecond)
		handler.writeUnauthorized(w)
		return
	}

	knownPass, ok := handler.userMap[user]
	if !ok {
		att.numAttemtps++
		att.timestamp = time.Now()
		time.Sleep(time.Duration(sleepTime) * time.Microsecond)
		handler.writeUnauthorized(w)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(knownPass), []byte(pass))
	if err != nil {
		att.numAttemtps++
		att.timestamp = time.Now()
		time.Sleep(time.Duration(sleepTime) * time.Microsecond)
		handler.writeUnauthorized(w)
		return
	}

	att.numAttemtps = 0
	handler.wrappedHandler.ServeHTTP(w, r)
}

func NewBasicAuthHandler(realm string, userMap map[string]string, handler http.Handler) BasicAuthHandler {
	var h BasicAuthHandler
	h.realm = realm
	h.wrappedHandler = handler
	h.userMap = userMap
	h.attemptMap = make(map[string]*attempt)
	h.rnd = rand.New(rand.NewSource(99))
	return h
}
