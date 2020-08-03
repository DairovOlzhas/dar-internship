// Package discussion provides the logic of the chat and all its components.
package discussion

import (
	"time"
)

// Discussion represents conversation between two people or group of peoples.
type Discussion struct {
	ID                 int64     `json:"id,omitempty"`
	FromId             string    `json:"from_id,omitempty"`
	ToId               string    `json:"to_id,omitempty"`
	CourseId           int64     `json:"course_id,omitempty"`
	IsActive           bool      `json:"is_active"`
	UnreadMessagesCnt  int64     `json:"unread_messages_cnt"`
	RecipientFirstName string    `json:"recipient_first_name,omitempty"`
	RecipientLastName  string    `json:"recipient_last_name,omitempty"`
	CourseName         string    `json:"course_name,omitempty"`
	Time               time.Time `json:"time,omitempty"`
}

// Update is the same as Discussion but have only changeable fields
// of Discussion instance.
type Update struct {
	UnreadMessagesCnt *int64     `json:"unread_messages_cnt:omitempty"`
	IsActive          *bool      `json:"is_active,omitempty"`
	Time              *time.Time `json:"time,omitempty"`
}

// Message represents information about sent messages between users
type Message struct {
	ID         int64     `json:"id"`
	FromId     string    `json:"from_id,omitempty"`
	ToId       string    `json:"to_id,omitempty"`
	CourseId   int64     `json:"course_id,omitempty"`
	Text       string    `json:"text,omitempty"`
	FilePath   string    `json:"file_path,omitempty"`
	IsRead     bool      `json:"is_read"`
	SentTime   time.Time `json:"sent_time,omitempty"`
	SenderName string    `json:"sender_name,omitempty"`
}

// UpdateMessage is the same as Message but have only changeable fields
// of Message instance.
type UpdateMessage struct {
	IsRead     *bool      `json:"is_read"`
	Text       *string    `json:"text,omitempty"`
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
