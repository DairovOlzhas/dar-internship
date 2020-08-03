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
		var userId string
		var ok bool
		userId, ok = r.Context().Value("user_id").(string)
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
		var userId string
		var ok bool
		userId, ok = r.Context().Value("user_id").(string)
		if !ok {
			userId, ok = mux.Vars(r)["user_id"]
			if !ok {
				return errSys.BadRequest(320, "No user ID")
			}
		}
		discussionIdStr, ok := mux.Vars(r)[idParam]
		if !ok {
			return errSys.BadRequest(320, "No discussion ID")
		}
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return errSys.Forbidden(10, "Websocket not supported: "+err.Error())
		}
		discussionId, err := strconv.ParseInt(discussionIdStr, 10, 64)
		if err != nil {
			return errSys.BadRequest(320, "Invalid discussion ID: "+err.Error())
		}
		cmd := &CommandStartDiscussion{userId, discussionId, conn}
		_, err = fac.ch.ExecCommand(cmd)
		if err != nil {
			return errSys.BadRequest(320, "Invalid command: "+err.Error())
		}
		return htp.OK(nil)
	}
}

func (fac *HttpEndpointFactory) GetDiscussionWith(recipientIdParam string) htp.Endpoint {
	return func(w http.ResponseWriter, r *http.Request) htp.Response {
		userId, ok := r.Context().Value("user_id").(string)
		if !ok {
			return errSys.BadRequest(320, "No user ID.")
		}
		recipientId, ok := mux.Vars(r)[recipientIdParam]
		if !ok {
			return errSys.BadRequest(320, "No recipient ID.")
		}
		cmd := &GetDiscussionWith{}
		cmd.UserID = userId
		cmd.RecipientID = recipientId
		resp, err := fac.ch.ExecCommand(cmd)
		if err != nil {
			return errSys.BadRequest(320, "Invalid command: "+err.Error())
		}
		return htp.Created(resp)
	}
}

func (fac *HttpEndpointFactory) DeleteDiscussion(idParam string) htp.Endpoint {
	return func(w http.ResponseWriter, r *http.Request) htp.Response {
		discussionIdStr, ok := mux.Vars(r)[idParam]
		if !ok {
			return errSys.BadRequest(320, "No discussion ID")
		}
		discussionId, err := strconv.ParseInt(discussionIdStr, 10, 64)
		if err != nil {
			return errSys.BadRequest(310, "Invalid ID: "+err.Error())
		}
		cmd := &CommandDeleteDiscussion{discussionId}
		resp, err := fac.ch.ExecCommand(cmd)
		if err != nil {
			return errSys.BadRequest(320, "Invalid command: "+err.Error())
		}
		return htp.OK(resp)
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

func (fac *HttpEndpointFactory) GetMessages(idParam string) htp.Endpoint {
	return func(w http.ResponseWriter, r *http.Request) htp.Response {
		userId, ok := r.Context().Value("user_id").(string)
		if !ok {
			return errSys.BadRequest(320, "No user ID.")
		}
		discussionIdStr, ok := mux.Vars(r)[idParam]
		if !ok {
			return errSys.BadRequest(320, "No discussion ID.")
		}
		discussionId, err := strconv.ParseInt(discussionIdStr, 10, 64)
		if err != nil {
			return errSys.BadRequest(320, "Invalid id:"+err.Error())
		}
		params, err := htp.ParseListParams(r)
		cmd := &CommandGetMessages{
			UserID:       userId,
			DiscussionID: discussionId,
			Params:       params,
		}
		resp, err := fac.ch.ExecCommand(cmd)
		if err != nil {
			return errSys.BadRequest(320, "Invalid command: "+err.Error())
		}
		return htp.OK(resp)
	}
}

func (fac *HttpEndpointFactory) FindByID(idParam string) htp.Endpoint {
	return func(w http.ResponseWriter, r *http.Request) htp.Response {
		userId, ok := r.Context().Value("user_id").(string)
		if !ok {
			return errSys.BadRequest(320, "No user ID.")
		}
		discussionIdStr, ok := mux.Vars(r)[idParam]
		if !ok {
			return errSys.BadRequest(320, "No discussion ID.")
		}
		discussionId, err := strconv.ParseInt(discussionIdStr, 10, 64)
		if err != nil {
			return errSys.BadRequest(320, "Invalid id: "+err.Error())
		}
		cmd := &CommandGetDiscussion{}
		cmd.UserID = userId
		cmd.DiscussionID = discussionId
		resp, err := fac.ch.ExecCommand(cmd)
		if err != nil {
			return errSys.BadRequest(320, "Invalid command: "+err.Error())
		}
		return htp.OK(resp)
	}
}
