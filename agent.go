package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"strings"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	logger "github.com/snail007/mini-logger"
)

type ErrorCode int

const (
	E_PARSE       ErrorCode = -32700
	E_INVALID_REQ ErrorCode = -32600
	E_NO_METHOD   ErrorCode = -32601
	E_BAD_PARAMS  ErrorCode = -32602
	E_INTERNAL    ErrorCode = -32603
	E_SERVER      ErrorCode = -32000
)

var ErrNullResult = errors.New("result is null")

type RPCError struct {
	// A Number that indicates the error type that occurred.
	Code ErrorCode `json:"code"` /* required */

	// A String providing a short description of the error.
	// The message SHOULD be limited to a concise single sentence.
	Message string `json:"message"` /* required */

	// A Primitive or Structured value that contains additional information about the error.
	Data interface{} `json:"data"` /* optional */
}

func (e *RPCError) Error() string {
	return e.Message
}

var null = json.RawMessage([]byte("null"))

// ----------------------------------------------------------------------------
// Request and Response
// ----------------------------------------------------------------------------

// serverRequest represents a ProtoRPC request received by the server.
type jsonRequest struct {
	// JSON-RPC protocol.
	Version string `json:"jsonrpc"`
	// A String containing the name of the method to be invoked.
	Method string `json:"method"`
	// An Array of objects to pass as arguments to the method.
	Params *json.RawMessage `json:"params"`
	// The request id. This can be of any type. It is used to match the
	// response with the request that it is replying to.
	Id *json.RawMessage `json:"id"`
}

// serverResponse represents a ProtoRPC response returned by the server.
type jsonResponse struct {
	// JSON-RPC protocol.
	Version string `json:"jsonrpc"`
	// The Object that was returned by the invoked method. This must be null
	// in case there was an error invoking the method.
	Result interface{} `json:"result"`
	// An Error object if there was an error invoking the method. It must be
	// null if there was no error.
	Error interface{} `json:"error"`
	// This must be the same id as the request it is responding to.
	Id *json.RawMessage `json:"id"`
}

var (
	services = new(serviceMap)
)

func call(jsonBytes []byte) (r jsonRequest, w jsonResponse, jsonResponseString string) {
	e := new(RPCError)
	defer func() {
		str, err := json.Marshal(w)
		if err != nil {
			e.Message = err.Error()
			e.Code = E_INTERNAL
			w.Error = e
			return
		}
		jsonResponseString = string(str)
	}()
	w.Version = "2.0"
	err := json.Unmarshal(jsonBytes, &r)
	if err != nil {
		e.Message = err.Error()
		e.Code = E_PARSE
		w.Error = e
		return
	}
	if r.Version != "2.0" {
		e.Message = "protocol error , only 2.0"
		e.Code = E_INVALID_REQ
		w.Error = e
		return
	}
	w.Id = r.Id
	w.Version = r.Version
	// Get service method to be called.
	serviceSpec, methodSpec, errGet := services.get(r.Method)
	if errGet != nil {
		e.Message = errGet.Error()
		e.Code = E_NO_METHOD
		w.Error = e
		return
	}
	reply := reflect.New(methodSpec.replyType)
	errValue := []reflect.Value{}

	if methodSpec.argsType != nil {
		if r.Params == nil {
			e.Message = "bad params"
			e.Code = E_BAD_PARAMS
			w.Error = e
			return
		}
		args := reflect.New(methodSpec.argsType)
		err1 := json.Unmarshal(*r.Params, args.Interface())
		if err1 != nil {
			e.Message = err1.Error()
			e.Code = E_BAD_PARAMS
			w.Error = e
			return
		}
		errValue = methodSpec.method.Func.Call([]reflect.Value{
			serviceSpec.rcvr,
			args,
			reply,
		})
	} else {
		errValue = methodSpec.method.Func.Call([]reflect.Value{
			serviceSpec.rcvr,
			reply,
		})
	}

	// Cast the result to error if needed.
	var errResult error
	errInter := errValue[0].Interface()
	if errInter != nil {
		errResult = errInter.(error)
	}

	if errResult == nil {
		w.Result = reply.Interface()
	} else {
		e.Message = errResult.Error()
		e.Code = E_INTERNAL
		w.Error = e
		return
	}
	return
}

func serve(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if !auth(w, r, ps) {
		id := json.RawMessage([]byte("0"))
		response := createErrorResponse(&id, E_INVALID_REQ, "auth fail", nil)
		body, e := json.Marshal(response)
		if e == nil {
			fmt.Fprint(w, string(body))
			return
		}
		fmt.Fprint(w, "auth fail")
		return
	}
	if isWS(r) {
		serveWS(w, r, ps)
	} else {
		serveHTTP(w, r, ps)
	}
}
func parseRequest(body []byte) (err error) {
	r := new(jsonRequest)
	err = json.Unmarshal(body, &r)
	return
}
func createErrorResponse(id *json.RawMessage, errorcode ErrorCode, errmsg string, result interface{}) (response jsonResponse) {
	r := new(jsonResponse)
	r.Id = id
	r.Version = "2.0"
	r.Error = RPCError{
		Code:    errorcode,
		Message: errmsg,
		Data:    result,
	}
	response = *r
	return
}
func isWS(r *http.Request) bool {
	if value := r.Header.Get("Upgrade"); strings.ToLower(value) != "websocket" {
		return false
	}
	if value := r.Header.Get("Connection"); strings.ToLower(value) != "upgrade" {
		return false
	}
	if value := r.Header.Get("Sec-WebSocket-Version"); strings.ToLower(value) != "13" {
		return false
	}
	return true
}
func auth(w http.ResponseWriter, r *http.Request, ps httprouter.Params) bool {
	return true
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func serveWS(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			c.WriteMessage(mt, message)
			if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) {
				//log.Warn(err)
			}
			break
		}
		_, _, j := call(message)
		err = c.WriteMessage(mt, []byte(j))
		if err != nil {
			log.With(logger.Fields{"uri": r.RequestURI, "addr": r.RemoteAddr}).Warn("write:", err)
			break
		}
	}
}
func serveHTTP(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "POST,OPTIONS")
	if r.Method == "OPTIONS" {
		return
	}
	result, err := ioutil.ReadAll(r.Body)
	if err == nil {
		_, _, j := call(result)
		fmt.Fprint(w, j)
	} else {
		fmt.Fprint(w, err.Error())
	}
}
