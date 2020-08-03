package discussion

import htp "git.dar.tech/dareco-go/http"

type Repository interface {
	FindAll(query FindQuery) ([]*Discussion, error)
	FindByID(id int64, userId string) (*Discussion, error)
	Create(*Discussion) (*Discussion, error)
	Update(id int64, upd *Update) error
	Delete(id int64) error

	FindAllMessages(params *htp.ListParams) ([]*Message, error)
	FindMessageByID(int64) (*Message, error)
	CreateMessage(*Message) (*Message, error)
	UpdateMessage(id int64, upd *UpdateMessage) error
	DeleteMessages(discussionId int64) error

	CreateFile(*File) (*File, error)

	GetParticipants(discussionId int64) ([]*Participant, error)
	AddParticipant(participant *Participant) error
	RemoveParticipant(participant *Participant) error
	DecrementUnreadMessagesCnt(participant *Participant) error
	IncrementUnreadMessagesCnt(participant *Participant) error
}
