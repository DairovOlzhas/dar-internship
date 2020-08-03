package discussion

// ConnGroup stores all clients of certain user
type ConnGroup struct {
	// all client of user
	clients    map[*Client]bool
	// broadcast is a channel which will store messages to broadcast
	broadcast  chan interface{}
	// register is a channel which will store clients to register
	register   chan *Client
	// unregister is a channel which will store clients to unregister
	unregister chan *Client
}

// Hub stores Connection Groups.
type Hub struct {
	// each ConnGroup belongs to single user of group discussion.
	// Respectively, there are two types of key:
	//  1) 'u' + user_id
	//	2) 'l' + course_id
	connGroups map[string]*ConnGroup
}

func NewHub() *Hub {
	return &Hub{
		map[string]*ConnGroup{},
	}
}

// GetOrCreateDiscussionHub returns or creates ConnGroup if it not exists.
func (hub *Hub) GetOrCreateConnGroup(connGroupId string) *ConnGroup {
	if chat, ok := hub.connGroups[connGroupId]; ok {
		return chat
	}
	hub.connGroups[connGroupId] = &ConnGroup{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan interface{}, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	go hub.run(connGroupId)
	return hub.connGroups[connGroupId]
}

// run starts loop where messages distributed to certain ConnGroup
func (hub *Hub) run(connGroupId string) {
	if connGroup, ok := hub.connGroups[connGroupId]; ok {
		for {
			select {
			case client := <-connGroup.register:
				connGroup.clients[client] = true
			case client := <-connGroup.unregister:
				for _, allConnGroups := range hub.connGroups {
					if _, ok := allConnGroups.clients[client]; ok {
						delete(allConnGroups.clients, client)
					}
				}
				close(client.Send)
			case message := <-connGroup.broadcast:
				for client := range connGroup.clients {
					select {
					case client.Send <- message:
					default:
						close(client.Send)
						delete(connGroup.clients, client)
					}
				}
			}
			for key, connGroup := range hub.connGroups {
				if len(connGroup.clients) == 0 {
					delete(hub.connGroups, key)
				}
			}
		}
	}
}

// SendTo sends message to specified ConnGroups' broadcasts.
func (hub *Hub) SendTo(msg interface{}, connGroupIds ...string) {
	for _, connGroupId := range connGroupIds {
		if connGroup, ok := hub.connGroups[connGroupId]; ok {
			connGroup.broadcast <- msg
		}
	}
}
