package discussion

import (
	"github.com/gorilla/websocket"
)

type FindQuery struct {
	DiscussionID *int64
	SenderID     *string
	RecipientID  *string
	IsGroup	 	 *bool
}

type CommandGetDiscussion struct {
	UserID string
	DiscussionID int64
}

func (cmd *CommandGetDiscussion) Exec(svc interface{}) (interface{}, error) {
	return svc.(Service).FindByID(cmd.DiscussionID, cmd.UserID)
}

type CommandStartDiscussions struct {
	UserId string          `json:"user_id"`
	Conn   *websocket.Conn `json:"-"`
}

func (cmd *CommandStartDiscussions) Exec(svc interface{}) (interface{}, error) {
	if err := svc.(Service).StartDiscussions(cmd.UserId, cmd.Conn); err != nil {
		return nil, err
	}
	return cmd, nil
}

type CommandStartDiscussion struct {
	UserId       string          `json:"user_id"`
	DiscussionId int64           `json:"discussion_id"`
	Conn         *websocket.Conn `json:"-"`
}

func (cmd *CommandStartDiscussion) Exec(svc interface{}) (interface{}, error) {
	if err := svc.(Service).StartDiscussion(cmd.UserId, cmd.DiscussionId, cmd.Conn); err != nil {
		return nil, err
	}
	return cmd, nil
}

//type CommandGetUserDiscussions struct {
//	UserId string
//}
//
//func (cmd *CommandGetUserDiscussions) Exec(svc interface{}) (interface{}, error) {
//	return svc.(Service).GetUserDiscussions(cmd.UserId)
//}

type GetDiscussionWith struct {
	UserID      string
	RecipientID string
}

func (cmd *GetDiscussionWith) Exec(svc interface{}) (interface{}, error) {
	return svc.(Service).GetDiscussionWith(cmd.UserID, cmd.RecipientID)
}

type CommandDeleteDiscussion struct {
	ID int64 `json:"id"`
}

func (cmd *CommandDeleteDiscussion) Exec(svc interface{}) (interface{}, error) { // TODO check permission
	if err := svc.(Service).Delete(cmd.ID); err != nil {
		return nil, err
	}
	return cmd, nil
}

//type  CommandMakeInactive struct {
//	DiscussionId int64
//}
//
//func (cmd *CommandMakeInactive) Exec(svc interface{}) (interface{}, error) {
//	return nil, svc.(Service).MakeInactive(cmd.DiscussionId)
//}
//
//type  CommandMakeActive struct {
//	DiscussionId int64
//}
//
//func (cmd *CommandMakeActive) Exec(svc interface{}) (interface{}, error) {
//	return nil, svc.(Service).MakeActive(cmd.DiscussionId)
//}

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

type CommandGetMessages struct {
	UserID       string
	DiscussionID int64
	Params       *htp.ListParams
}

func (cmd *CommandGetMessages) Exec(svc interface{}) (interface{}, error) {
	participants, err := svc.(Service).GetParticipants(cmd.DiscussionID)
	if err != nil {
		return nil, err
	}
	for _, p := range participants {
		if p.ID == cmd.UserID {
			cmd.Params.Query["discussion_id"] = cmd.DiscussionID
			return svc.(Service).FindAllMessages(cmd.Params)
		}
	}
	return nil, ErrNoPermission
}

//type CommandSendViolation struct {
//	SenderId     string
//	DiscussionId int64
//	Text         string
//}
//
//func (cmd *CommandSendViolation) Exec(svc interface{}) (interface{}, error) {
//	err := svc.(Service).SendViolation(&Violation{
//		SenderId:     cmd.SenderId,
//		DiscussionId: cmd.DiscussionId,
//		Text:         cmd.Text,
//	})
//	if err != nil {
//		return nil, err
//	}
//	return nil, nil
//}
//
//type CommandGetViolations struct {
//	params *http.ListParams
//}
//
//func (cmd *CommandGetViolations) Exec(svc interface{}) (interface{}, error) {
//	return svc.(Service).GetViolations(cmd.params)
//}

//type CommandBlockOrUnblockUser struct {
//	RequestingUser       string
//	DiscussionId 		 int64
//	Block 				 bool
//}
//
//func (cmd *CommandBlockOrUnblockUser) Exec(svc interface{}) (interface{},error) {
//	if cmd.Block {
//		err := svc.(Service).BlockUser(cmd.RequestingUser, cmd.DiscussionId)
//		if err != nil {
//			return nil, err
//		}
//	} else {
//		err := svc.(Service).UnblockUser(cmd.RequestingUser, cmd.DiscussionId)
//		if err != nil {
//			return nil, err
//		}
//	}
//	return nil, nil
//}
