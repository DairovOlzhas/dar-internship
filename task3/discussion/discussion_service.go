package discussion

import (
	"git.dar.tech/dareco-go/utils/uuid"
	"git.dar.tech/education/tutor/uploader"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"strconv"
	"sync"
)

// Service implements
type Service interface {
	GetUserDiscussions(userId string) ([]*Discussion, error)
	CreateDiscussion(discussion *Discussion) (*Discussion, error)
	DeleteDiscussion(id int64) error

	StartDiscussions(userId string, conn *websocket.Conn) error
	StartDiscussion(userId string, discussionId int64, conn *websocket.Conn) error

	AddFile(*FileRequest) (*File, error)
}

type svc struct {
	hub      *Hub
	repo     Repository
	uploader uploader.Uploader
}

// NewService creates and returns Service interface.
func NewService(hub *Hub, repo Repository, uploader uploader.Uploader) Service {
	return &svc{
		hub:      hub,
		repo:     repo,
		uploader: uploader,
	}
}

func (svc *svc) GetUserDiscussions(userId string) ([]*Discussion, error) {
	return svc.repo.FindAll(FindQuery{FromId: &userId}, nil)
}

func (svc *svc) CreateDiscussion(discussion *Discussion) (*Discussion, error) {
	if discussion.FromId == "" || (discussion.ToId == "" && discussion.CourseId == 0) {
		return nil, ErrInvalidDiscussion
	}
	err := svc.repo.Create(discussion)
	if err != nil {
		return nil, err
	}
	discussion, err = svc.repo.FindByID(discussion.ID)
	if err != nil {
		return nil, err
	}
	svc.hub.SendTo(discussion, "u"+discussion.FromId)
	if discussion.ToId != "" {
		rDiscussion := &Discussion{
			FromId:   discussion.ToId,
			ToId:     discussion.FromId,
			CourseId: discussion.CourseId,
			Time:     discussion.Time,
		}
		err := svc.repo.Create(rDiscussion)
		if err != nil {
			return nil, err
		}
		rDiscussion, err = svc.repo.FindByID(rDiscussion.ID)
		if err != nil {
			return nil, err
		}
		svc.hub.SendTo(rDiscussion, "u"+rDiscussion.FromId)
	}
	return discussion, nil
}

func (svc *svc) DeleteDiscussion(id int64) error {
	discussion, err := svc.repo.FindByID(id)
	if err != nil {
		return err
	}
	err = svc.repo.DeleteMessages(discussion.FromId, discussion.ToId, discussion.CourseId)
	if err != nil {
		return err
	}
	discussionsToDelete := []*Discussion{}

	discussions, err := svc.repo.FindAll(FindQuery{FromId: &discussion.FromId, ToId: &discussion.ToId}, nil)
	discussionsToDelete = append(discussionsToDelete, discussions...)

	discussions, err = svc.repo.FindAll(FindQuery{FromId: &discussion.ToId, ToId: &discussion.FromId}, nil)
	discussionsToDelete = append(discussionsToDelete, discussions...)

	if discussion.CourseId != 0 {
		discussions, err = svc.repo.FindAll(FindQuery{CourseId: &discussion.CourseId}, nil)
		discussionsToDelete = append(discussionsToDelete, discussions...)
	}
	err = svc.repo.DeleteMessages(discussion.FromId, discussion.ToId, discussion.CourseId)
	if err != nil {
		return err
	}
	for _, discussion := range discussionsToDelete {
		svc.hub.SendTo(discussionID(discussion.ID), "u"+discussion.FromId)
		err := svc.repo.Delete(discussion.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (svc *svc) StartDiscussions(userId string, conn *websocket.Conn) error {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	client := &Client{
		ConnGroup: svc.hub.GetOrCreateConnGroup("u" + userId),
		WsConn:    conn,
		Send:      make(chan interface{}, 256),
	}
	client.DataTypes = []DataType{
		&typeMessage{svc, userId},
		&typeDiscussion{svc, userId, client},
		&typeDeleteDiscussion{},
		&typeReadMessages{userId, svc},
	}
	defer func() {
		if !client.connClosed {
			err := client.WsConn.Close()
			if err != nil {
				log.Errorf("Can't close websocket connection: %v", err)
			}
			client.connClosed = true
		}
	}()
	client.ConnGroup.register <- client
	discussions, err := svc.GetUserDiscussions(userId)
	if err != nil {
		return err
	}

	var err1, err2 error
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		err1 = client.WritePump(wg)
	}()
	for _, msg := range discussions {
		if msg.CourseId != 0 {
			connGroup := svc.hub.GetOrCreateConnGroup("l" + strconv.FormatInt(msg.CourseId, 10))
			connGroup.register <- client
		}
		client.Send <- msg
	}
	go func() {
		err2 = client.ReadPump(wg)
	}()
	wg.Wait()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}

func (svc *svc) StartDiscussion(userId string, discussionId int64, conn *websocket.Conn) error {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	client := &Client{
		ConnGroup: svc.hub.GetOrCreateConnGroup("u" + userId),
		WsConn:    conn,
		Send:      make(chan interface{}, 256),
	}
	client.DataTypes = []DataType{
		&typeMessage{svc, userId},
		&typeDiscussion{svc, userId, client},
	}
	defer func() {
		if !client.connClosed {
			err := client.WsConn.Close()
			if err != nil {
				log.Errorf("Can't close websocket connection: %v", err)
			}
			client.connClosed = true
		}
	}()

	client.ConnGroup.register <- client
	discussion, err := svc.repo.FindByID(discussionId)
	if err != nil {
		return err
	}

	var err1, err2 error
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		err1 = client.WritePump(wg)
	}()
	pastMessages, err := svc.repo.FindAllMessages(
		FindQuery{
			FromId:   &discussion.FromId,
			ToId:     &discussion.ToId,
			CourseId: &discussion.CourseId,
		})
	if err != nil {
		return err
	}
	for _, message := range pastMessages {
		client.Send <- message
	}
	go func() {
		err2 = client.ReadPump(wg)
	}()
	wg.Wait()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}

	return nil
}

func (svc *svc) AddFile(fileRequest *FileRequest) (*File, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	path, err := svc.uploader.Upload(fileRequest.File, "/messages/files/"+id.String()+"."+fileRequest.Extension)
	if err != nil {
		return nil, err
	}
	return svc.repo.CreateFile(&File{fileRequest.OwnerID, path})
}
