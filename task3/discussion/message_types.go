package discussion

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"
)

type typeMessage struct {
	svc    *svc
	userId string
}

func (t *typeMessage) IsMyInstance(data []byte, instance interface{}) (interface{}, bool) {
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

func (t *typeMessage) WriteProcess(instance interface{}) *DataPacket {
	return &DataPacket{
		"message",
		instance.(*Message),
	}
}

func (t *typeMessage) ReadProcess(instance interface{}) error {
	message := instance.(*Message)

	if (message.ToId != "" && message.CourseId != 0) ||
		(message.Text == "" && message.FilePath == "") ||
		(message.CourseId == 0 && message.ToId == "") {
		return ErrInvalidMessage
	}

	message.FromId = t.userId
	message.IsRead = false
	message.SentTime = time.Now()

	message, err := t.svc.repo.CreateMessage(message)
	if err != nil {
		return err
	}
	message, err = t.svc.repo.FindMessageByID(message.ID)
	if err != nil {
		return err
	}

	if message.CourseId != 0 { // group chat
		err := t.svc.repo.ExecQuery(
			`UPDATE tutor_discussions SET unread_messages_cnt = unread_messages_cnt + 1 
					WHERE course_id = $1 AND from_id != $2`,
			message.CourseId,
			message.FromId,
		)
		if err != nil {
			return err
		}
		t.svc.hub.SendTo(message, "l"+strconv.FormatInt(message.CourseId, 10))
		t.svc.hub.SendTo(&Discussion{CourseId: message.CourseId, Time: message.SentTime}, "l"+strconv.FormatInt(message.CourseId, 10))
	} else { // one-to-one chat
		err := t.svc.repo.ExecQuery(
			`UPDATE tutor_discussions SET unread_messages_cnt = unread_messages_cnt + 1
					WHERE from_id = $1 AND to_id = $2`,
			message.ToId,
			message.FromId,
		)
		if err != nil {
			return err
		}
		err = t.svc.repo.ExecQuery(
			`UPDATE tutor_discussions SET time = $1 
					WHERE (from_id = $2 AND to_id = $3) OR (from_id = $3 AND to_id = $2)`,
			message.SentTime,
			message.FromId,
			message.ToId,
		)
		if err != nil {
			return err
		}

		discussionsToSend := []*Discussion{}
		discussions, err := t.svc.repo.FindAll(FindQuery{FromId: &message.FromId, ToId: &message.ToId}, nil)
		if err != nil {
			return err
		}
		discussionsToSend = append(discussionsToSend, discussions...)
		discussions, err = t.svc.repo.FindAll(FindQuery{FromId: &message.ToId, ToId: &message.FromId}, nil)
		if err != nil {
			return err
		}
		discussionsToSend = append(discussionsToSend, discussions...)

		t.svc.hub.SendTo(message, "u"+t.userId, "u"+message.ToId)
		for _, discussion := range discussionsToSend {
			t.svc.hub.SendTo(discussion, "u"+discussion.FromId)
		}
	}
	return nil
}

type typeDiscussion struct {
	svc    *svc
	userId string
	client *Client
}

func (t typeDiscussion) IsMyInstance(data []byte, instance interface{}) (interface{}, bool) {
	if data != nil { // read
		var discussionId int64
		discussionPacket := DataPacket{Data: &discussionId}
		if json.Unmarshal(data, &discussionPacket) == nil && discussionPacket.Type == "discussion" {
			return discussionId, true
		}
	} else if instance != nil { // write
		if discussion, ok := instance.(*Discussion); ok {
			return discussion, true
		}
	}
	return nil, false
}

func (t typeDiscussion) WriteProcess(instance interface{}) *DataPacket {
	discussion := instance.(*Discussion)
	discussionPacket := &DataPacket{}
	discussionPacket.Type = "discussion"
	if discussion.CourseId != 0 {
		discussion.FromId = ""
	}
	discussionPacket.Data = discussion
	return discussionPacket
}

func (t typeDiscussion) ReadProcess(instance interface{}) error {
	discussionId := instance.(int64)
	discussion, err := t.svc.repo.FindByID(discussionId)
	if err != nil {
		return err
	}
	if discussion.CourseId == 0 && discussion.ToId == "" {
		return ErrNoRecipient
	}
	discussion.FromId = t.userId
	pastMessages, err := t.svc.repo.FindAllMessages(
		FindQuery{
			FromId:   &discussion.FromId,
			ToId:     &discussion.ToId,
			CourseId: &discussion.CourseId,
		})
	if err != nil {
		return err
	}
	for _, message := range pastMessages {
		t.client.Send <- message
	}
	return nil
}

type typeDeleteDiscussion struct{}

type discussionID int64

func (t typeDeleteDiscussion) IsMyInstance(data []byte, instance interface{}) (interface{}, bool) {
	if instance != nil { // write
		if discussionID, ok := instance.(discussionID); ok {
			return discussionID, true
		}
	}
	return nil, false
}

func (t typeDeleteDiscussion) WriteProcess(instance interface{}) *DataPacket {
	return &DataPacket{
		"delete_discussion",
		instance.(discussionID),
	}
}

func (t typeDeleteDiscussion) ReadProcess(i interface{}) error {
	return nil
}

type typeReadMessages struct {
	userId string
	svc    *svc
}

type messageID int64

func (t *typeReadMessages) IsMyInstance(data []byte, instance interface{}) (interface{}, bool) {
	if data != nil { // read
		var messageId int64
		readMessagePacket := DataPacket{Data: &messageId}
		if json.Unmarshal(data, &readMessagePacket) == nil && readMessagePacket.Type == "read_message" {
			return messageId, true
		}
	} else if instance != nil { // write
		if messageID, ok := instance.(messageID); ok {
			return messageID, true
		}
	}
	return nil, false
}

func (t *typeReadMessages) WriteProcess(instance interface{}) *DataPacket {
	return &DataPacket{
		"read_message",
		instance.(messageID),
	}
}

func (t *typeReadMessages) ReadProcess(instance interface{}) error {
	messageId := instance.(int64)
	message, err := t.svc.repo.FindMessageByID(messageId)
	if err != nil {
		return err
	}
	if message.CourseId != 0 {
		return errors.New("Reading group messages unsupported.")
	}
	if message.IsRead {
		return ErrReadMessage
	}
	if t.userId == message.FromId {
		return ErrReadOwnMessage
	}
	err = t.svc.repo.ExecQuery(
		`UPDATE tutor_discussion_messages SET is_read = true WHERE id = $1`,
		messageId,
	)
	if err != nil {
		return err
	}
	t.svc.hub.SendTo(messageID(messageId), "u"+message.FromId)

	err = t.svc.repo.ExecQuery(
		`UPDATE tutor_discussions SET unread_messages_cnt = unread_messages_cnt - 1
					WHERE from_id = $1 AND to_id = $2`,
		message.ToId,
		message.FromId,
	)
	if err != nil {
		return err
	}

	discussions, err := t.svc.repo.FindAll(FindQuery{FromId: &message.ToId, ToId: &message.FromId}, nil)
	if err != nil {
		return err
	}

	for _, discussion := range discussions {
		t.svc.hub.SendTo(discussion, "u"+discussion.FromId)
	}

	return nil
}
