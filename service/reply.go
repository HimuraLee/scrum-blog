package service

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/pkg/errors"
	"net/http"
)

type Reply struct {
	Code int         `json:"code"`
	Msg  interface{} `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func newReply(code int, msg, data interface{}) (int, Reply) {
	var reply Reply
	reply.Code = code
	reply.Msg = msg
	if code != http.StatusOK {
		if errEcho, ok := data.(echo.HTTPError); ok {
			logrus.Error(msg, "; error: ", errEcho.Error())
			reply.Data = errEcho.Message
		} else if errSimple, ok := data.(error); ok {
			logrus.Error(msg, "; error: ", errSimple)
			reply.Data = errSimple.Error()
		} else if s, ok := data.(string); ok {
			logrus.Error(msg, "; error: ", s)
			reply.Data = s
		} else {
			logrus.Error(msg, "; error: ", data)
			reply.Data = data
		}
	}
	if data != nil {
		reply.Data = data
	}
	return 200, reply
}

func Suc(msg interface{}, data ...interface{}) (int, Reply) {
	if len(data) > 0 {
		return newReply(http.StatusOK, msg, data[0])
	}
	return newReply(http.StatusOK, msg, nil)
}

func BadRequest(msg interface{}, data ...interface{}) (int, Reply) {
	if len(data) > 0 {
		return newReply(http.StatusBadRequest, msg, data[0])
	}
	return newReply(http.StatusBadRequest, msg, nil)
}

func NotFound(msg interface{}, data ...interface{}) (int, Reply) {
	if len(data) > 0 {
		return newReply(http.StatusNotFound, msg, data[0])
	}
	return newReply(http.StatusNotFound, msg, nil)
}

func AuthFailed(msg interface{}, data ...interface{}) (int, Reply) {
	if len(data) > 0 {
		return newReply(http.StatusForbidden, msg, data[0])
	}
	return newReply(http.StatusForbidden, msg, nil)
}

func Unauthorized(msg interface{}, data ...interface{}) (int, Reply) {
	if len(data) > 0 {
		return newReply(http.StatusUnauthorized, msg, data[0])
	}
	return newReply(http.StatusUnauthorized, msg, nil)
}

func InternalError(msg interface{}, data ...interface{}) (int, Reply) {
	if len(data) > 0 {
		return newReply(http.StatusInternalServerError, msg, data[0])
	}
	return newReply(http.StatusInternalServerError, msg, nil)
}

func CheckResponse(resp *http.Response) (*Reply, error) {
	var (
		reply Reply
		err   error
	)
	d := json.NewDecoder(resp.Body)
	if err = d.Decode(&reply); err != nil {
		return nil, err
	}
	logrus.Info("reply recv", "reply", reply)
	if reply.Code != http.StatusOK {
		if reply.Msg != nil {
			return &reply, errors.New(reply.Msg.(string))
		}
		return &reply, errors.New("unexpected error")
	}
	return &reply, nil
}
