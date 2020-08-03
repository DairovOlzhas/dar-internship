package discussion

import (
	"git.dar.tech/dareco-go/http"
)

// Repository interacts with database and makes queries related to this package
type Repository interface {
	FindAll(query FindQuery, params *http.PaginationParams) ([]*Discussion, error)
	FindByID(int64) (*Discussion, error)
	Create(*Discussion) error
	Update(id int64, upd *Update) error
	Delete(id int64) error

	FindAllMessages(query FindQuery) ([]*Message, error)
	FindMessageByID(int64) (*Message, error)
	CreateMessage(*Message) (*Message, error)
	UpdateMessage(id int64, upd *UpdateMessage) error
	DeleteMessages(fromId string, toId string, courseId int64) error

	CreateFile(*File) (*File, error)

	ExecQuery(q string, values ...interface{}) error
}
