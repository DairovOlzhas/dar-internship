package discussion

import (
	"github.com/gorilla/websocket"
	"time"
)

type FindQuery struct {
	FromId             *string
	NotFromId          *string
	ToId               *string
	CourseId           *int64
	IsActive           *bool
	IsRead             *bool
	SentTime           *time.Time
	Time               *time.Time
	RecipientFirstName *string
	RecipientLastName  *string
	CourseName         *string
}

// CommandStartDiscussions starts all discussions of user
// by calling StartDiscussions method of Service.
//
// You should provide UserId and Conn.
type CommandStartDiscussions struct {
	UserId string
	Conn   *websocket.Conn
}

func (cmd *CommandStartDiscussions) Exec(svc interface{}) (interface{}, error) {
	return nil, svc.(Service).StartDiscussions(cmd.UserId, cmd.Conn)
}

// CommandStartDiscussions starts single discussion of user
// by calling StartDiscussion method of Service.
//
// You should provide UserId, DiscussionId and Conn.
type CommandStartDiscussion struct {
	UserId       string
	DiscussionId int64
	Conn         *websocket.Conn
}

func (cmd *CommandStartDiscussion) Exec(svc interface{}) (interface{}, error) {
	return nil, svc.(Service).StartDiscussion(cmd.UserId, cmd.DiscussionId, cmd.Conn)
}

// CommandGetUserDiscussions returns all discussion of user
// by calling GetUserDiscussions method of Service.
//
// You should provide UserId.
type CommandGetUserDiscussions struct {
	UserId string
}

func (cmd *CommandGetUserDiscussions) Exec(svc interface{}) (interface{}, error) {
	return svc.(Service).GetUserDiscussions(cmd.UserId)
}

// CommandGetUserDiscussions creates discussion for user
// by calling CreateDiscussion method of Service.
//
// You should provide Discussion.
type CommandCreateDiscussion struct {
	Discussion *Discussion
}

func (cmd *CommandCreateDiscussion) Exec(svc interface{}) (interface{}, error) {
	discussion, err := svc.(Service).CreateDiscussion(cmd.Discussion)
	if err != nil {
		return nil, err
	}
	return discussion, err
}

// CommandDeleteDiscussion deletes discussion of user
// by calling DeleteDiscussion method of Service.
//
// You should provide UserId and DiscussionId.
type CommandDeleteDiscussion struct {
	UserId       string
	DiscussionId int64
}

func (cmd *CommandDeleteDiscussion) Exec(svc interface{}) (interface{}, error) {
	return nil, svc.(Service).DeleteDiscussion(cmd.DiscussionId)
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