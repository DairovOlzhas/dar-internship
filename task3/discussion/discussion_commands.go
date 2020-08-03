package discussion

import (
	htp "git.dar.tech/dareco-go/http"
	"github.com/gorilla/websocket"
)

type FindQuery struct {
	DiscussionID *int64
	SenderID     *string
	RecipientID  *string
	IsGroup	 	 *bool
}

type CommandGetDiscussion struct {
	UserID string
	DiscussionID int64
}

func (cmd *CommandGetDiscussion) Exec(svc interface{}) (interface{}, error) {
	return svc.(Service).FindByID(cmd.DiscussionID, cmd.UserID)
}

// CommandStartDiscussions starts all discussions of user
// by calling StartDiscussions method of Service.
//
// You should provide UserId and Conn.
type CommandStartDiscussions struct {
	UserId string          `json:"user_id"`
	Conn   *websocket.Conn `json:"-"`
}

func (cmd *CommandStartDiscussions) Exec(svc interface{}) (interface{}, error) {
	if err := svc.(Service).StartDiscussions(cmd.UserId, cmd.Conn); err != nil {
		return nil, err
	}
	return cmd, nil
}

// CommandStartDiscussion starts single discussion of user
// by calling StartDiscussion method of Service.
//
// You should provide UserId, DiscussionId and Conn.
type CommandStartDiscussion struct {
	UserId       string          `json:"user_id"`
	DiscussionId int64           `json:"discussion_id"`
	Conn         *websocket.Conn `json:"-"`
}

func (cmd *CommandStartDiscussion) Exec(svc interface{}) (interface{}, error) {
	if err := svc.(Service).StartDiscussion(cmd.UserId, cmd.DiscussionId, cmd.Conn); err != nil {
		return nil, err
	}
	return cmd, nil
}

// CommandGetUserDiscussions returns all discussion of user
// by calling GetUserDiscussions method of Service.
type CommandGetUserDiscussions struct {
	UserId string
}

func (cmd *CommandGetUserDiscussions) Exec(svc interface{}) (interface{}, error) {
	return svc.(Service).GetUserDiscussions(cmd.UserId)
}

type GetDiscussionWith struct {
	UserID      string
	RecipientID string
}

func (cmd *GetDiscussionWith) Exec(svc interface{}) (interface{}, error) {
	return svc.(Service).GetDiscussionWith(cmd.UserID, cmd.RecipientID)
}

// CommandDeleteDiscussion deletes discussion of user
// by calling DeleteDiscussion method of Service.
//
// You should provide ID.
type CommandDeleteDiscussion struct {
	ID int64 `json:"id"`
}

func (cmd *CommandDeleteDiscussion) Exec(svc interface{}) (interface{}, error) { // TODO check permission
	if err := svc.(Service).Delete(cmd.ID); err != nil {
		return nil, err
	}
	return cmd, nil
}

// CommandAddFile uploads received file to S3 and
// return link to uploaded file by calling AddFile method of Service.
//
// You should provide OwnerID, Extension and File.
type CommandAddFile struct {
	OwnerID   string
	Extension string
	File      []byte
}

func (cmd *CommandAddFile) Exec(svc interface{}) (interface{}, error) {
	file := &FileRequest{
		OwnerID:   cmd.OwnerID,
		Extension: cmd.Extension,
		File:      cmd.File,
	}
	res, err := svc.(Service).AddFile(file)
	if err != nil {
		return nil, err
	}
	return res, nil
}

type CommandGetMessages struct {
	UserID       string
	DiscussionID int64
	Params       *htp.ListParams
}

func (cmd *CommandGetMessages) Exec(svc interface{}) (interface{}, error) {
	participants, err := svc.(Service).GetParticipants(cmd.DiscussionID)
	if err != nil {
		return nil, err
	}
	for _, p := range participants {
		if p.ID == cmd.UserID {
			cmd.Params.Query["discussion_id"] = cmd.DiscussionID
			return svc.(Service).FindAllMessages(cmd.Params)
		}
	}
	return nil, ErrNoPermission
}
