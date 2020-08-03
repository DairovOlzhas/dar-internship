// Package discussion provides the logic of the chat and all its components.
package discussion

import (
	"time"
)

// Discussion represents conversation between two people or group of peoples.
type Discussion struct {
	ID                int64        `json:"id,omitempty"`
	IsActive          bool         `json:"is_active"`
	IsGroup           bool         `json:"is_group"`
	Name              string       `json:"name,omitempty"`
	Photo             string       `json:"photo,omitempty"`
	Recipient         *Participant `json:"recipient,omitempty"`
	UnreadMessagesCnt int64        `json:"unread_messages_cnt"`
	SenderID		  string 	   `json:"sender_id,omitempty"`
	Time              time.Time    `json:"time,omitempty"`
	ParticipantsIds   []string     `json:"participant_ids,omitempty"`
}

// Update is the same as Discussion but have only changeable fields
// of Discussion instance.
type Update struct {
	Name     *string    `json:"name,omitempty"`
	Photo    *string    `json:"photo,omitempty"`
	IsActive *bool      `json:"is_active,omitempty"`
	Time     *time.Time `json:"time,omitempty"`
}

// Message represents information about sent messages between users
type Message struct {
	ID           int64        `json:"id"`
	Sender       *Participant `json:"sender"`
	DiscussionID int64        `json:"discussion_id"`
	Text         string       `json:"text,omitempty"`
	FilePath     string       `json:"file_path,omitempty"`
	IsRead       bool         `json:"is_read"`
	SentTime     time.Time    `json:"sent_time,omitempty"`
}

// UpdateMessage is the same as Message but have only changeable fields
// of Message instance.
type UpdateMessage struct {
	IsRead *bool   `json:"is_read"`
	Text   *string `json:"text,omitempty"`
}

type Participant struct {
	ID           string `json:"id,omitempty"`
	DiscussionID int64  `json:"-"`
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	Photo        string `json:"photo,omitempty"`
}

type ParticipantUpdate struct {
	UnreadMessagesCnt *int64
}

// FileRequest needs for receiving files and
// uploading them to cloud native storage.
type FileRequest struct {
	OwnerID   string `json:"owner_id,omitempty"`
	Extension string `json:"extension,omitempty"`
	File      []byte `json:"file,omitempty"`
}

// File is the response for file uploading request
type File struct {
	OwnerID  string `json:"owner_id,omitempty"`
	FilePath string `json:"file_path,omitempty"`
}