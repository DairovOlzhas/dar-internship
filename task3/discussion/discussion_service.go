package discussion

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Service interface {
	GetUserDiscussions(userId string) ([]*Discussion, error)
	GetDiscussionWith(senderId, recipientId string) (*Discussion, error)
	FindByID(id int64, userId string) (*Discussion, error)
	Create(discussion *Discussion) (*Discussion, error)
	Delete(id int64) error
	Update(id int64, upd *Update) error

	StartDiscussions(userId string, conn *websocket.Conn) error
	StartDiscussion(userId string, discussionId int64, conn *websocket.Conn) error

	AddFile(*FileRequest) (*File, error)


	GetParticipants(discussionId int64) ([]*Participant, error)
	AddParticipant(participant *Participant) error
	RemoveParticipant(participant *Participant) error


	FindAllMessages(params *htp.ListParams) ([]*Message, error)
	FindMessageByID(int64) (*Message, error)
	CreateMessage(*Message) (*Message, error)
	UpdateMessage(id int64, upd *UpdateMessage) error
	DeleteMessages(discussionId int64) error

	DecrementUnreadMessagesCnt(participant *Participant) error
	IncrementUnreadMessagesCnt(participant *Participant) error
}

type svc struct {
	hub      *Hub
	repo     Repository
	userRepo user.Repository
	uploader uploader.Uploader
}

func NewService(hub *Hub, repo Repository, uploader uploader.Uploader, userRepo user.Repository) Service {
	return &svc{
		hub:      hub,
		repo:     repo,
		userRepo: userRepo,
		uploader: uploader,
	}
}

func (svc *svc) GetUserDiscussions(userId string) ([]*Discussion, error) {
	return svc.repo.FindAll(FindQuery{SenderID: &userId})
}

func (svc *svc) GetDiscussionWith(userId, recipientId string) (*Discussion, error) {
	isGroup := false
	discussions, err := svc.repo.FindAll(
		FindQuery{
			SenderID:    &userId,
			RecipientID: &recipientId,
			IsGroup:     &isGroup,
		},
	)
	if err != nil {
		return nil, err
	}
	if len(discussions) == 1 {
		return discussions[0], nil
	}
	discussion := &Discussion{
		ParticipantsIds: []string{userId, recipientId},
	}
	discussion, err = svc.Create(discussion)
	if err != nil {
		return nil, err
	}
	return svc.FindByID(discussion.ID, userId)
}

func (svc *svc) FindByID(id int64, userId string) (*Discussion, error) {
	return svc.repo.FindByID(id, userId)
}

func (svc *svc) Create(discussion *Discussion) (*Discussion, error) {
	participantIds := map[string]bool{}
	for _, participantId := range discussion.ParticipantsIds {
		participantIds[participantId] = true
	}
	if !discussion.IsGroup && len(participantIds) != 2 || discussion.IsGroup && discussion.Name == "" {
		return nil, ErrInvalidDiscussion
	}

	discussion, err := svc.repo.Create(discussion)
	if err != nil {
		return nil, err
	}
	for participantId, _ := range participantIds {
		u, err := svc.userRepo.FindByID(participantId)
		if err != nil {
			return nil, err
		}
		participant := &Participant{
			ID:           u.ID,
			DiscussionID: discussion.ID,
			FirstName:    u.FirstName,
			LastName:     u.LastName,
			Photo:        u.Photo,
		}
		err = svc.repo.AddParticipant(participant)
		if err != nil {
			return nil, err
		}
	}
	discussions, err := svc.repo.FindAll(FindQuery{
		DiscussionID: &discussion.ID,
	})
	for _, d := range discussions {
		svc.hub.SendToUserClients(d.SenderID, d)
	}
	return discussion, nil
}

func (svc *svc) Delete(id int64) error {
	svc.hub.SendToDiscussion(id, DeleteDiscussionID(id))
	return svc.repo.Delete(id)
}

func (svc *svc) Update(id int64, upd *Update) error {
	return svc.repo.Update(id, upd)
}

func (svc *svc) StartDiscussions(userId string, conn *websocket.Conn) error {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	client := NewClientConnection(
		conn,
		NewMessageEvent(svc.hub, svc, userId),
		NewDiscussionEvent(),
		NewDeleteDiscussionEvent(),
		NewReadMessageEvent(svc.hub, svc, userId),
	)
	defer svc.hub.Unregister(userId, client)
	svc.hub.RegisterInHub(userId, client)
	discussions, err := svc.GetUserDiscussions(userId)
	if err != nil {
		return err
	}
	for _, discussion := range discussions {
		svc.hub.RegisterInDiscussion(userId, discussion.ID, client)
	}
	if err := client.Start(); err != nil {
		return err
	}
	return nil
}

func (svc *svc) StartDiscussion(userId string, discussionId int64, conn *websocket.Conn) error {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	client := NewClientConnection(
		conn,
		NewMessageEvent(svc.hub, svc, userId),
		NewDiscussionEvent(),
		NewDeleteDiscussionEvent(),
		NewReadMessageEvent(svc.hub, svc, userId),
	)
	defer svc.hub.Unregister(userId, client)
	discussion, err := svc.FindByID(discussionId, userId)
	if err != nil {
		return err
	}
	svc.hub.RegisterInDiscussion(userId, discussion.ID, client)
	if err := client.Start(); err != nil {
		return err
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

func (svc *svc) FindAllMessages(params *htp.ListParams) ([]*Message, error) {
	return svc.repo.FindAllMessages(params)
}

func (svc *svc) FindMessageByID(id int64) (*Message, error) {
	return svc.repo.FindMessageByID(id)
}

func (svc *svc) CreateMessage(message *Message) (*Message, error) {
	return svc.repo.CreateMessage(message)
}

func (svc *svc) UpdateMessage(id int64, upd *UpdateMessage) error {
	return svc.repo.UpdateMessage(id, upd)
}

func (svc *svc) DeleteMessages(discussionId int64) error {
	return svc.repo.DeleteMessages(discussionId)
}

func (svc *svc) GetParticipants(discussionId int64) ([]*Participant, error) {
	return svc.repo.GetParticipants(discussionId)
}

func (svc *svc) AddParticipant(participant *Participant) error {
	return svc.repo.AddParticipant(participant)
}

func (svc *svc) RemoveParticipant(participant *Participant) error {
	return svc.repo.RemoveParticipant(participant)
}

func (svc *svc) DecrementUnreadMessagesCnt(participant *Participant) error {
	return svc.repo.DecrementUnreadMessagesCnt(participant)
}

func (svc *svc) IncrementUnreadMessagesCnt(participant *Participant) error {
	return svc.repo.IncrementUnreadMessagesCnt(participant)
}
