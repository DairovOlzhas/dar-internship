package discussion

import (
	"encoding/json"
)

type Event interface {
	IsMyInstance(data []byte, instance interface{}) (interface{}, bool)
	WriteProcess(instance interface{}) *DataPacket
	ReadProcess(instance interface{}) error
}

type messageEvent struct {
	hub           *Hub
	svc           Service
	currentUserId string
}

func NewMessageEvent(hub *Hub, svc Service, currentUserId string) Event {
	return &messageEvent{
		hub:           hub,
		svc:           svc,
		currentUserId: currentUserId,
	}
}

func (t *messageEvent) IsMyInstance(data []byte, instance interface{}) (interface{}, bool) {
	if data != nil { // reading data
		message := &Message{}
		messagePacket := DataPacket{Data: message}
		if json.Unmarshal(data, &messagePacket) == nil && messagePacket.Type == "message" {
			return message, true
		}
	} else if instance != nil { // writing data
		if message, ok := instance.(*Message); ok {
			return message, true
		}
	}
	return nil, false
}

func (t *messageEvent) WriteProcess(instance interface{}) *DataPacket {
	return &DataPacket{
		"message",
		instance.(*Message),
	}
}

func (t *messageEvent) ReadProcess(instance interface{}) error {
	message := instance.(*Message)

	if (message.DiscussionID == 0) ||
		(message.Text == "" && message.FilePath == "") {
		return ErrInvalidMessage
	}
	message.IsRead = false
	message.Sender = &Participant{ID: t.currentUserId}
	message, err := t.svc.CreateMessage(message)
	if err != nil {
		return err
	}
	message, err = t.svc.FindMessageByID(message.ID)
	if err != nil {
		return err
	}

	participants, err := t.svc.GetParticipants(message.DiscussionID)
	if err != nil {
		return err
	}

	err = t.svc.Update(message.DiscussionID, &Update{Time: &message.SentTime})
	if err != nil {
		return err
	}

	t.hub.SendToDiscussion(message.DiscussionID, message)
	for _, participant := range participants {
		if participant.ID != t.currentUserId {
			err := t.svc.IncrementUnreadMessagesCnt(participant)
			if err != nil {
				return err
			}
		}
		discussion, err := t.svc.FindByID(message.DiscussionID, participant.ID)
		if err != nil {
			return err
		}
		t.hub.SendToUserClientsInDiscussion(discussion.ID, discussion.SenderID, discussion)
	}
	return nil
}

type discussionEvent struct{}

func NewDiscussionEvent() Event {
	return &discussionEvent{}
}

func (t discussionEvent) IsMyInstance(data []byte, instance interface{}) (interface{}, bool) {
	if instance != nil { // write
		if discussion, ok := instance.(*Discussion); ok {
			return discussion, true
		}
	}
	return nil, false
}

func (t discussionEvent) WriteProcess(instance interface{}) *DataPacket {
	return &DataPacket{
		Type: "discussion",
		Data: instance.(*Discussion),
	}
}

func (t discussionEvent) ReadProcess(instance interface{}) error {
	return nil
}

type deleteDiscussionEvent struct{}

func NewDeleteDiscussionEvent() Event {
	return &deleteDiscussionEvent{}
}

type DeleteDiscussionID int64

func (t deleteDiscussionEvent) IsMyInstance(data []byte, instance interface{}) (interface{}, bool) {
	if instance != nil { // write
		if discussionID, ok := instance.(DeleteDiscussionID); ok {
			return discussionID, true
		}
	}
	return nil, false
}

func (t deleteDiscussionEvent) WriteProcess(instance interface{}) *DataPacket {
	return &DataPacket{
		"delete_discussion",
		instance.(DeleteDiscussionID),
	}
}

func (t deleteDiscussionEvent) ReadProcess(i interface{}) error {
	return nil
}

type readMessageEvent struct {
	currentUserId string
	svc           Service
	hub           *Hub
}

func NewReadMessageEvent(hub *Hub, svc Service, currentUserId string) Event {
	return &readMessageEvent{
		hub:           hub,
		currentUserId: currentUserId,
		svc:           svc,
	}
}

type ReadMessageID int64

func (t *readMessageEvent) IsMyInstance(data []byte, instance interface{}) (interface{}, bool) {
	if data != nil { // read
		var messageId int64
		readMessagePacket := DataPacket{Data: &messageId}
		if json.Unmarshal(data, &readMessagePacket) == nil && readMessagePacket.Type == "read_message" {
			return messageId, true
		}
	} else if instance != nil { // write
		if messageID, ok := instance.(ReadMessageID); ok {
			return messageID, true
		}
	}
	return nil, false
}

func (t *readMessageEvent) WriteProcess(instance interface{}) *DataPacket {
	return &DataPacket{
		"read_message",
		instance.(ReadMessageID),
	}
}

func (t *readMessageEvent) ReadProcess(instance interface{}) error {
	messageId := instance.(int64)
	message, err := t.svc.FindMessageByID(messageId)
	if err != nil {
		return err
	}
	if message.IsRead {
		return ErrReadMessage
	}
	if t.currentUserId == message.Sender.ID {
		return ErrReadOwnMessage
	}
	isRead := true
	err = t.svc.UpdateMessage(message.ID, &UpdateMessage{IsRead: &isRead})
	if err != nil {
		return err
	}

	t.hub.SendToDiscussion(message.DiscussionID, ReadMessageID(message.ID))
	participants, err := t.svc.GetParticipants(message.DiscussionID)
	for _, participant := range participants {
		if participant.ID != message.Sender.ID {
			err := t.svc.DecrementUnreadMessagesCnt(participant)
			if err != nil {
				return err
			}
			discussion, err := t.svc.FindByID(participant.DiscussionID, participant.ID)
			if err != nil {
				return err
			}
			t.hub.SendToUserClientsInDiscussion(discussion.ID, discussion.SenderID, discussion)
		}
	}
	return nil
}
