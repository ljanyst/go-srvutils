//------------------------------------------------------------------------------
// Author: Lukasz Janyst <lukasz@jany.st>
// Date: 07.09.2019
//
// Licensed under the MIT License, see the LICENSE file for details.
//------------------------------------------------------------------------------

package auth

import (
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type BasicAuthHandler struct {
	userMap        map[string]string
	realm          string
	wrappedHandler http.Handler
}

func (handler BasicAuthHandler) writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, handler.realm))
	w.WriteHeader(401)
	w.Write([]byte("Unauthorised.\n"))
}

func (handler BasicAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if !ok {
		handler.writeUnauthorized(w)
		return
	}

	knownPass, ok := handler.userMap[user]
	if !ok {
		handler.writeUnauthorized(w)
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(knownPass), []byte(pass))
	if err != nil {
		handler.writeUnauthorized(w)
		return
	}

	handler.wrappedHandler.ServeHTTP(w, r)
}

func NewBasicAuthHandler(realm string, userMap map[string]string, handler http.Handler) BasicAuthHandler {
	var h BasicAuthHandler
	h.realm = realm
	h.wrappedHandler = handler
	h.userMap = userMap
	return h
}
