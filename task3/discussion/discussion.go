package discussion

import (
	"time"
)

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

type Update struct {
	Name     *string    `json:"name,omitempty"`
	Photo    *string    `json:"photo,omitempty"`
	IsActive *bool      `json:"is_active,omitempty"`
	Time     *time.Time `json:"time,omitempty"`
}

type Message struct {
	ID           int64        `json:"id"`
	Sender       *Participant `json:"sender"`
	DiscussionID int64        `json:"discussion_id"`
	Text         string       `json:"text,omitempty"`
	FilePath     string       `json:"file_path,omitempty"`
	IsRead       bool         `json:"is_read"`
	SentTime     time.Time    `json:"sent_time,omitempty"`
}

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

type FileRequest struct {
	OwnerID   string `json:"owner_id,omitempty"`
	Extension string `json:"extension,omitempty"`
	File      []byte `json:"file,omitempty"`
}

type File struct {
	OwnerID  string `json:"owner_id,omitempty"`
	FilePath string `json:"file_path,omitempty"`
}

type Violation struct {
	ID              int    `json:"id,omitempty"`
	SenderId        string `json:"sender_id,omitempty"`
	SenderFirstName string `json:"sender_first_name,omitempty"`
	SenderLastName  string `json:"sender_last_name,omitempty"`
	DiscussionId    int    `json:"discussion_id,omitempty"`
	Text            string `json:"text,omitempty"`
}
