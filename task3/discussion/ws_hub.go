package discussion

type discussionHub struct {
	userClients    map[string]map[ClientConnection]bool
	broadcast      chan interface{}
}

type Hub struct {
	discussionHubs map[int64]*discussionHub
	userClients    map[string]map[ClientConnection]bool
}

func NewHub() *Hub {
	return &Hub{
		discussionHubs: map[int64]*discussionHub{},
		userClients:    map[string]map[ClientConnection]bool{},
	}
}

func (hub *Hub) getOrCreateDiscussionHub(discussionId int64) *discussionHub {
	if chat, ok := hub.discussionHubs[discussionId]; ok {
		return chat
	}
	hub.discussionHubs[discussionId] = &discussionHub{
		userClients:    map[string]map[ClientConnection]bool{},
		broadcast:      make(chan interface{}, 256),
	}
	go hub.run(discussionId)
	return hub.discussionHubs[discussionId]
}

func (hub *Hub) run(discussionId int64) {
	if dHub, ok := hub.discussionHubs[discussionId]; ok {
		for {
			select {
			case message, ok := <-dHub.broadcast:
				if !ok {
					return
				}
				for userId := range dHub.userClients {
					for client := range dHub.userClients[userId] {
						select {
						case client.DataChan() <- message:
						default:
							close(client.DataChan())
							delete(dHub.userClients[userId], client)
						}
					}
				}
			}
		}
	}
}

// SendToDiscussion sends message to all clients in this discussion
func (hub *Hub) SendToDiscussion(discussionId int64, msg interface{}) {
	if dHub, ok := hub.discussionHubs[discussionId]; ok {
		dHub.broadcast <- msg
	}
}

// SendToUserClientsInDiscussion sends message to all user's clients in this discussion
func (hub *Hub) SendToUserClientsInDiscussion(discussionId int64, userId string, msg interface{}) {
	if _, ok := hub.discussionHubs[discussionId]; ok {
		userClientsInDiscussion := hub.discussionHubs[discussionId].userClients
		if _, ok := userClientsInDiscussion[userId]; ok {
			for userClient := range userClientsInDiscussion[userId] {
				userClient.DataChan() <- msg
			}
		}
	}
}

// SendToUserClientsInDiscussion sends message to all user's clients in hub
func (hub *Hub) SendToUserClients(userId string, msg interface{}) {
	if _, ok := hub.userClients[userId]; ok {
		for userClient := range hub.userClients[userId] {
			userClient.DataChan() <- msg
		}
	}
}

// RegisterInHub adds a client to the hub
func (hub *Hub) RegisterInHub(userId string, client ClientConnection) {
	if _, ok := hub.userClients[userId]; !ok {
		hub.userClients[userId] = map[ClientConnection]bool{}
	}
	hub.userClients[userId][client] = true
}

// Register adds a client to the discussion hub
func (hub *Hub) RegisterInDiscussion(userId string, discussionId int64, client ClientConnection) {
	dHub := hub.getOrCreateDiscussionHub(discussionId)
	if _, ok := dHub.userClients[userId]; !ok {
		dHub.userClients[userId] = map[ClientConnection]bool{}
	}
	dHub.userClients[userId][client] = true
}

// Unregister removes the client of user from each discussion hub if client exists in it
func (hub *Hub) Unregister(userId string, client ClientConnection) {
	for discussionId, dHub := range hub.discussionHubs {
		if _, ok := dHub.userClients[userId]; ok {
			if _, ok := dHub.userClients[userId][client]; ok {
				// removing user's client from discussion hub
				delete(dHub.userClients[userId], client)
			}
			if len(dHub.userClients[userId]) == 0 {
				// removing user from discussion hub if he doesn't have any clients
				delete(dHub.userClients, userId)
			}
		}
		if len(dHub.userClients) == 0 {
			delete(hub.discussionHubs, discussionId)
			close(dHub.broadcast)
		}
	}
	if _, ok := hub.userClients[userId]; ok {
		if _, ok := hub.userClients[userId][client]; ok {
			// removing user's client from hub
			delete(hub.userClients[userId], client)
		}
		if len(hub.userClients[userId]) == 0 {
			// removing user from hub if he doesn't have any clients
			delete(hub.userClients, userId)
		}
	}
	close(client.DataChan())
}
