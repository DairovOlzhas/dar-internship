package discussion

import (
	"git.dar.tech/dareco-go/cqrses"
	htp "git.dar.tech/dareco-go/http"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"net/http"
	"strconv"
)

var (
	errSys = htp.ErrorSystem{System: "DISCUSSION", Series: 10}

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)


type HttpEndpointFactory struct {
	ch cqrses.CommandHandler
}

func NewEndpointFactory(ch cqrses.CommandHandler) *HttpEndpointFactory {
	return &HttpEndpointFactory{ch}
}

func (fac *HttpEndpointFactory) StartDiscussionsWs() htp.Endpoint {
	return func(w http.ResponseWriter, r *http.Request) htp.Response {
		userId, ok := r.Context().Value("user_id").(string)
		if !ok {
			userId, ok = mux.Vars(r)["user_id"]
			if !ok {
				return errSys.BadRequest(320, "No user ID")
			}
		}
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return errSys.Forbidden(10, "Websocket not supported: "+err.Error())
		}
		cmd := &CommandStartDiscussions{userId, conn}
		_, err = fac.ch.ExecCommand(cmd)
		if err != nil {
			return errSys.BadRequest(320, "Invalid command: "+err.Error())
		}
		return htp.OK(nil)
	}
}

func (fac *HttpEndpointFactory) StartDiscussionWs(idParam string) htp.Endpoint {
	return func(w http.ResponseWriter, r *http.Request) htp.Response {
		userId, ok := r.Context().Value("user_id").(string)
		if !ok {
			userId, ok = mux.Vars(r)["user_id"]
			if !ok {
				return errSys.BadRequest(320, "No user ID")
			}
		}

		discussionId, ok := mux.Vars(r)[idParam]
		if !ok {
			return errSys.BadRequest(320, "No discussion ID")
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return errSys.Forbidden(10, "Websocket not supported: "+err.Error())
		}
		id, err := strconv.ParseInt(discussionId, 10, 64)
		if err != nil {
			return errSys.BadRequest(320, "Invalid discussion ID: "+err.Error())
		}
		cmd := &CommandStartDiscussion{userId, id, conn}
		_, err = fac.ch.ExecCommand(cmd)
		if err != nil {
			return errSys.BadRequest(320, "Invalid command: "+err.Error())
		}
		return htp.OK(nil)
	}
}

func (fac *HttpEndpointFactory) GetUserDiscussions() htp.Endpoint {
	return func(w http.ResponseWriter, r *http.Request) htp.Response {
		userId, ok := r.Context().Value("user_id").(string)
		if !ok {
			return errSys.BadRequest(320, "No user ID")
		}
		cmd := &CommandGetUserDiscussions{userId}
		resp, err := fac.ch.ExecCommand(cmd)
		if err != nil {
			return errSys.BadRequest(320, "Invalid command: "+err.Error())
		}
		return htp.OK(resp)
	}
}

func (fac *HttpEndpointFactory) CreateDiscussion() htp.Endpoint {
	return func(w http.ResponseWriter, r *http.Request) htp.Response {
		_, ok := r.Context().Value("user_id").(string)
		if !ok {
			return errSys.BadRequest(320, "No user ID")
		}
		var discussion *Discussion
		err := htp.ParseBody(r, &discussion)
		if err != nil {
			return errSys.BadRequest(310, "Invalid json: "+err.Error())
		}
		cmd := &CommandCreateDiscussion{discussion}
		resp, err := fac.ch.ExecCommand(cmd)
		if err != nil {
			return errSys.BadRequest(320, "Invalid command: "+err.Error())
		}
		return htp.Created(resp)
	}
}

func (fac *HttpEndpointFactory) DeleteDiscussion(idParam string) htp.Endpoint {
	return func(w http.ResponseWriter, r *http.Request) htp.Response {
		userId, ok := r.Context().Value("user_id").(string)
		if !ok {
			return errSys.BadRequest(320, "No user ID")
		}
		id, ok := mux.Vars(r)[idParam]
		if !ok {
			return errSys.BadRequest(320, "No discussion ID")
		}
		discussionId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return errSys.BadRequest(310, "Invalid ID: "+err.Error())
		}
		cmd := &CommandDeleteDiscussion{userId, discussionId}
		_, err = fac.ch.ExecCommand(cmd)
		if err != nil {
			return errSys.BadRequest(320, "Invalid command: "+err.Error())
		}
		return htp.OK(nil)
	}
}

func (fac *HttpEndpointFactory) AddFile() htp.Endpoint {
	return func(w http.ResponseWriter, r *http.Request) htp.Response {
		userId, ok := r.Context().Value("user_id").(string)
		if !ok {
			return errSys.BadRequest(320, "No user ID")
		}
		cmd := &CommandAddFile{}
		contentType := r.Header.Get("Content-Type")
		var extension string
		var data []byte
		switch contentType {
		case "image/gif":
			extension = "gif"
		case "image/jpeg":
			extension = "jpg"
		case "image/svg+xml":
			extension = "svg"
		case "application/pdf":
			extension = "pdf"
		case "application/doc":
			extension = "doc"
		default:
			return errSys.BadRequest(320, "File type not supported")
		}
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return errSys.BadRequest(320, "Invalid command: "+err.Error())
		}
		cmd.OwnerID = userId
		cmd.Extension = extension
		cmd.File = data
		resp, err := fac.ch.ExecCommand(cmd)
		if err != nil {
			return errSys.BadRequest(320, "Invalid command: "+err.Error())
		}
		return htp.OK(resp)
	}
}
