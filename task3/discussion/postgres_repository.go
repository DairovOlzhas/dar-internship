package discussion

import (
	"database/sql"
	"fmt"
	htp "git.dar.tech/dareco-go/http"
	"git.dar.tech/dareco-go/utils/postgres"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

type discussionRow struct {
	ID                 int64          `json:"id,omitempty"`
	IsActive           bool           `json:"is_active"`
	IsGroup            bool           `json:"is_group"`
	Name               sql.NullString `json:"name,omitempty"`
	Photo              sql.NullString `json:"photo,omitempty"`
	RecipientID        sql.NullString `json:"recipient_id,omitempty"`
	RecipientFirstName sql.NullString `json:"recipient_first_name,omitempty"`
	RecipientLastName  sql.NullString `json:"recipient_last_name,omitempty"`
	RecipientPhoto     sql.NullString `json:"recipient_photo,omitempty"`
	UnreadMessagesCnt  int64          `json:"unread_messages_cnt"`
	SenderID           string         `json:"-"`
	Time               time.Time      `json:"time,omitempty"`
}

func (item *discussionRow) toDiscussion() *Discussion {
	discussion := &Discussion{
		ID:                item.ID,
		IsActive:          item.IsActive,
		IsGroup:           item.IsGroup,
		Name:              item.Name.String,
		Photo:             item.Photo.String,
		UnreadMessagesCnt: item.UnreadMessagesCnt,
		SenderID:          item.SenderID,
		Time:              item.Time,
	}
	if !item.IsGroup {
		discussion.Recipient = &Participant{
			ID:        item.RecipientID.String,
			FirstName: item.RecipientFirstName.String,
			LastName:  item.RecipientLastName.String,
			Photo:     item.RecipientPhoto.String,
		}
	}
	return discussion
}

type messageRow struct {
	ID              int64          `json:"id"`
	DiscussionID    int64          `json:"discussion_id"`
	Text            sql.NullString `json:"text,omitempty"`
	FilePath        sql.NullString `json:"file_path,omitempty"`
	IsRead          bool           `json:"is_read"`
	SentTime        time.Time      `json:"sent_time,omitempty"`
	SenderID        string         `json:"sender_id,omitempty"`
	SenderFirstName sql.NullString `json:"sender_first_name,omitempty"`
	SenderLastName  sql.NullString `json:"sender_last_name,omitempty"`
	SenderPhoto     sql.NullString `json:"sender_photo,omitempty"`
}

func (item *messageRow) toMessage() *Message {
	return &Message{
		ID:           item.ID,
		DiscussionID: item.DiscussionID,
		Text:         item.Text.String,
		FilePath:     item.FilePath.String,
		IsRead:       item.IsRead,
		SentTime:     item.SentTime,
		Sender: &Participant{
			ID:        item.SenderID,
			FirstName: item.SenderFirstName.String,
			LastName:  item.SenderLastName.String,
			Photo:     item.SenderPhoto.String,
		},
	}
}

type participantRow struct {
	ID           string         `json:"id,omitempty"`
	DiscussionID int64          `json:"discussion_id,omitempty"`
	FirstName    sql.NullString `json:"first_name,omitempty"`
	LastName     sql.NullString `json:"last_name,omitempty"`
	Photo        sql.NullString `json:"photo,omitempty"`
}

func (item *participantRow) toParticipant() *Participant {
	return &Participant{
		ID:           item.ID,
		DiscussionID: item.DiscussionID,
		FirstName:    item.FirstName.String,
		LastName:     item.LastName.String,
		Photo:        item.Photo.String,
	}
}

type discussionRepo struct {
	db                   *sql.DB
	discussionsTableName string
	violationsTable      string
}

func NewPostgresRepoWithDB(db *sql.DB) (Repository, error) {
	var discussionsTableName = "tutor_discussions"
	var violationsTable = "tutor_discussion_violations"
	var messageRepoQueries = []string{
		`create table tutor_discussions
(
    id        serial                                 not null
        constraint tutor_discussions_pkey
            primary key,
    is_active boolean                  default true,
    time      timestamp with time zone default now(),
    name      varchar(255),
    is_group  boolean                  default false not null,
    photo     varchar(255),
    constraint tutor_discussions_check
        check ((((name)::text <> ''::text) AND (name IS NOT NULL)) OR (NOT is_group))
);`,
		`create table tutor_discussion_messages
(
    id            serial                                 not null
        constraint tutor_discussion_messages_pkey
            primary key,
    text          text,
    file_path     text,
    is_read       boolean                  default false not null,
    sent_time     timestamp with time zone default now(),
    sender_id     varchar(255),
    discussion_id integer                                not null,
    constraint tutor_discussion_messages_sender_id_fkey
        foreign key (sender_id, discussion_id) references tutor_discussion_participants (user_id, discussion_id)
            on delete cascade
);`,
		`create table tutor_discussion_participants
(
    user_id             varchar(255) default 'deleted'::character varying not null
        constraint tutor_discussion_participants_user_id_fkey
            references tutor_users
            on delete set default,
    discussion_id       integer                                           not null
        constraint tutor_discussion_participants_discussion_id_fkey
            references tutor_discussions
            on delete cascade,
    unread_messages_cnt integer      default 0                            not null,
    constraint tutor_discussion_participants_pkey
        primary key (discussion_id, user_id)
);`, `
	create table tutor_discussion_files
(
    id        serial       not null
        constraint tutor_discussion_files_pkey
            primary key,
    owner_id  varchar(255) not null
        constraint tutor_discussion_files_owner_id_fkey
            references tutor_users,
    file_path varchar(255) not null
);
`,
	}
	var err error
	for _, q := range messageRepoQueries {
		_, err = db.Exec(q)
		if err != nil {
			log.Errorln(q, err)
		}
	}
	return &discussionRepo{
			db,
			discussionsTableName,
			violationsTable},
		nil
}

func NewPostgresRepo(cfg postgres.Config) (Repository, error) {
	db, err := postgres.NewDBSession(cfg)
	if err != nil {
		return nil, err
	}
	return NewPostgresRepoWithDB(db)
}

func (repo *discussionRepo) FindAll(query FindQuery) ([]*Discussion, error) {
	var items []*Discussion
	q := `SELECT
       discussions.id        AS id,
       discussions.is_active AS is_active,
       discussions.is_group  AS is_group,
       discussions.time      AS time,
       discussions.name      AS name,
       discussions.photo     AS photo,
       recipients.id         AS recipient_id,
       recipients.first_name AS recipient_first_name,
       recipients.last_name  AS recipient_last_name,
       recipients.photo      AS recipient_photo,
       participants.unread_messages_cnt AS unread_messages_cnt,
       participants.user_id AS sender_id
FROM tutor_discussions AS discussions
         LEFT JOIN tutor_discussion_participants AS participants
                   ON discussions.id = participants.discussion_id
         LEFT JOIN tutor_discussion_participants AS recipient_ids
                   ON discussions.id = recipient_ids.discussion_id
                          AND discussions.is_group = false
                          AND participants.user_id != recipient_ids.user_id
        LEFT JOIN tutor_users AS recipients
                    ON recipient_ids.user_id = recipients.id `

	cnt := 0
	values := []interface{}{}
	parts := []string{}
	if query.DiscussionID != nil {
		cnt++
		parts = append(parts, "participants.discussion_id = $"+strconv.Itoa(cnt))
		values = append(values, *query.DiscussionID)
	}
	if query.SenderID != nil {
		cnt++
		parts = append(parts, "participants.user_id = $"+strconv.Itoa(cnt))
		values = append(values, *query.SenderID)
	}
	if query.RecipientID != nil {
		cnt++
		parts = append(parts, "recipients.id = $"+strconv.Itoa(cnt))
		values = append(values, *query.RecipientID)
	}
	if query.IsGroup != nil {
		cnt++
		parts = append(parts, "discussions.is_group = $"+strconv.Itoa(cnt))
		values = append(values, *query.IsGroup)
	}
	if len(values) <= 0 {
		return []*Discussion{}, nil
	} else {
		q = q + ` WHERE `
	}

	q = q + strings.Join(parts, " AND ")
	rows, err := repo.db.Query(q, values...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		item := discussionRow{}
		err := rows.Scan(
			&item.ID,
			&item.IsActive,
			&item.IsGroup,
			&item.Time,
			&item.Name,
			&item.Photo,
			&item.RecipientID,
			&item.RecipientFirstName,
			&item.RecipientLastName,
			&item.RecipientPhoto,
			&item.UnreadMessagesCnt,
			&item.SenderID,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item.toDiscussion())
	}
	return items, nil
}

func (repo *discussionRepo) FindByID(id int64, userId string) (*Discussion, error) {
	item := &discussionRow{}
	q := `SELECT
       discussions.id        AS id,
       discussions.is_active AS is_active,
       discussions.is_group  AS is_group,
       discussions.time      AS time,
       discussions.name      AS name,
       discussions.photo     AS photo,
       recipients.id         AS recipient_id,
       recipients.first_name AS recipient_first_name,
       recipients.last_name  AS recipient_last_name,
       recipients.photo      AS recipient_photo,
       participants.unread_messages_cnt AS unread_messages_cnt,
       participants.user_id AS sender_id
FROM tutor_discussions AS discussions
         LEFT JOIN tutor_discussion_participants AS participants
                   ON discussions.id = participants.discussion_id
         LEFT JOIN tutor_discussion_participants AS recipient_ids
                   ON discussions.id = recipient_ids.discussion_id
                          AND discussions.is_group = false
                          AND participants.user_id != recipient_ids.user_id
        LEFT JOIN tutor_users AS recipients
                    ON recipient_ids.user_id = recipients.id 
WHERE participants.discussion_id = $1 AND participants.user_id = $2 `
	err := repo.db.QueryRow(q, id, userId).Scan(
		&item.ID,
		&item.IsActive,
		&item.IsGroup,
		&item.Time,
		&item.Name,
		&item.Photo,
		&item.RecipientID,
		&item.RecipientFirstName,
		&item.RecipientLastName,
		&item.RecipientPhoto,
		&item.UnreadMessagesCnt,
		&item.SenderID,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return item.toDiscussion(), nil
}

func (repo *discussionRepo) Create(item *Discussion) (*Discussion, error) {
	q := `INSERT INTO ` + repo.discussionsTableName +
		` (is_active, is_group, name, photo, time)
			VALUES ($1,$2,$3,$4,$5)
		RETURNING id,time`
	err := repo.db.QueryRow(q,
		item.IsActive,
		item.IsGroup,
		item.Name,
		item.Photo,
		time.Now(),
	).Scan(
		&item.ID,
		&item.Time,
	)
	if err != nil {
		return nil, ErrDiscussionCreation
	}
	return item, nil
}

func (repo *discussionRepo) Update(id int64, upd *Update) error {
	if id == 0 {
		return ErrUpdate
	}
	q := fmt.Sprintf("UPDATE %s SET ", repo.discussionsTableName)

	parts := []string{}
	values := []interface{}{}
	cnt := 0
	if upd.Time != nil {
		cnt++
		parts = append(parts, "time = $"+strconv.Itoa(cnt))
		values = append(values, *upd.Time)
	}
	if upd.IsActive != nil {
		cnt++
		parts = append(parts, "is_active = $"+strconv.Itoa(cnt))
		values = append(values, *upd.IsActive)
	}
	if upd.Name != nil {
		cnt++
		parts = append(parts, "name = $"+strconv.Itoa(cnt))
		values = append(values, *upd.Name)
	}
	if upd.Photo != nil {
		cnt++
		parts = append(parts, "photo = $"+strconv.Itoa(cnt))
		values = append(values, *upd.Photo)
	}
	if len(parts) <= 0 {
		return ErrNothingToUpdate
	}
	cnt++
	q = q + strings.Join(parts, " , ") + " WHERE id = $" + strconv.Itoa(cnt)
	values = append(values, id)

	stmt, err := repo.db.Prepare(q)
	if err != nil {
		return err
	}
	defer func() {
		if clsErr := stmt.Close(); clsErr != nil {
			log.Warn(clsErr)
		}
	}()

	ret, err := stmt.Exec(values...)
	if err != nil {
		return err
	}
	n, err := ret.RowsAffected()
	if err != nil {
		return err
	}
	if n <= 0 {
		return ErrNotFound
	}
	return nil
}

func (repo *discussionRepo) Delete(id int64) error {
	q := `DELETE FROM ` + repo.discussionsTableName + ` WHERE id = $1`
	ret, err := repo.db.Exec(q, id)
	if err != nil {
		return err
	}
	n, err := ret.RowsAffected()
	if err != nil {
		return err
	}
	if n <= 0 {
		return ErrNotFound
	}
	return nil
}

func (repo *discussionRepo) FindAllMessages(params *htp.ListParams) ([]*Message, error) {
	items := []*Message{}
	q := `SELECT
			messages.id,
			messages.discussion_id,
			messages.text,
			messages.file_path,
			messages.is_read,
			messages.sent_time,
			users.id,
			users.first_name,
			users.last_name,
			users.photo
		FROM tutor_discussion_messages AS messages
			LEFT JOIN tutor_users AS users
			ON messages.sender_id = users.id 
`
	cnt := 0
	parts := []string{}
	values := []interface{}{}
	if sentTime, ok := params.Query["sent_time"]; ok {
		cnt++
		parts = append(parts, "sent_time < $"+strconv.Itoa(cnt))
		values = append(values, sentTime)
	}
	if discussionId, ok := params.Query["discussion_id"]; ok {
		cnt++
		parts = append(parts, "discussion_id = $"+strconv.Itoa(cnt))
		values = append(values, discussionId)
	}
	if isRead, ok := params.Query["is_read"]; ok {
		cnt++
		parts = append(parts, "is_read = $"+strconv.Itoa(cnt))
		values = append(values, isRead)
	}
	if senderId, ok := params.Query["sender_id"]; ok {
		cnt++
		parts = append(parts, "users.id = $"+strconv.Itoa(cnt))
		values = append(values, senderId)
	}
	if len(values) > 0 {
		q = q + ` WHERE `
	} else {
		return nil, ErrMessageNotFound
	}
	q = q + strings.Join(parts, " AND ") + ` ORDER BY messages.sent_time ASC `

	if params.Limit() > 0 {
		q += fmt.Sprintf(" LIMIT %d", params.Limit())
	}
	if params.Pagination.Page > 0 {
		q += fmt.Sprintf(" OFFSET %d", params.Pagination.Page)
	}

	rows, err := repo.db.Query(q, values...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		item := messageRow{}
		err := rows.Scan(
			&item.ID,
			&item.DiscussionID,
			&item.Text,
			&item.FilePath,
			&item.IsRead,
			&item.SentTime,
			&item.SenderID,
			&item.SenderFirstName,
			&item.SenderLastName,
			&item.SenderPhoto,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item.toMessage())
	}
	for i:=0; i < len(items)/2; i++ {
		items[i], items[len(items)-i-1] = items[len(items)-i-1], items[i]
	}
	return items, nil
}

func (repo *discussionRepo) FindMessageByID(id int64) (*Message, error) {
	item := &messageRow{}
	q := `SELECT
			messages.id,
			messages.discussion_id,
			messages.text,
			messages.file_path,
			messages.is_read,
			messages.sent_time,
			users.id,
			users.first_name,
			users.last_name,
			users.photo
		FROM tutor_discussion_messages AS messages
			LEFT JOIN tutor_users AS users
			ON messages.sender_id = users.id
		WHERE messages.id = $1`

	err := repo.db.QueryRow(q, id).Scan(
		&item.ID,
		&item.DiscussionID,
		&item.Text,
		&item.FilePath,
		&item.IsRead,
		&item.SentTime,
		&item.SenderID,
		&item.SenderFirstName,
		&item.SenderLastName,
		&item.SenderPhoto,
	)
	if err == sql.ErrNoRows {
		return nil, ErrMessageNotFound
	} else if err != nil {
		return nil, err
	}
	return item.toMessage(), nil
}

func (repo *discussionRepo) CreateMessage(item *Message) (*Message, error) {
	q := `
		INSERT INTO tutor_discussion_messages
			(sender_id, discussion_id, text, file_path, is_read, sent_time) 
			VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, sent_time
	`
	err := repo.db.QueryRow(q,
		item.Sender.ID,
		item.DiscussionID,
		item.Text,
		item.FilePath,
		item.IsRead,
		time.Now(),
	).Scan(
		&item.ID,
		&item.SentTime,
	)
	if err != nil {
		return nil, ErrMessageCreation
	}
	return item, nil
}

func (repo *discussionRepo) UpdateMessage(id int64, upd *UpdateMessage) error {
	if id == 0 {
		return ErrUpdate
	}
	q := "UPDATE tutor_discussion_messages SET "

	parts := []string{}
	values := []interface{}{}
	cnt := 0
	if upd.IsRead != nil {
		cnt++
		parts = append(parts, "is_read = $"+strconv.Itoa(cnt))
		values = append(values, *upd.IsRead)
	}
	if upd.Text != nil {
		cnt++
		parts = append(parts, "time = $"+strconv.Itoa(cnt))
		values = append(values, *upd.Text)
	}
	if len(parts) <= 0 {
		return ErrNothingToUpdate
	}
	cnt++
	q = q + strings.Join(parts, " , ") + " WHERE id = $" + strconv.Itoa(cnt)
	values = append(values, id)

	stmt, err := repo.db.Prepare(q)
	if err != nil {
		return err
	}
	defer func() {
		if clsErr := stmt.Close(); clsErr != nil {
			log.Warn(clsErr)
		}
	}()

	ret, err := stmt.Exec(values...)
	if err != nil {
		return err
	}
	n, err := ret.RowsAffected()
	if err != nil {
		return err
	}
	if n <= 0 {
		return ErrMessageNotFound
	}
	return nil
}

func (repo *discussionRepo) DeleteMessages(discussionId int64) error {
	q := `DELETE FROM tutor_discussion_messages WHERE discussion_id = $1`
	_, err := repo.db.Query(q, discussionId)
	if err != nil {
		return err
	}
	return nil
}

func (repo *discussionRepo) CreateFile(item *File) (*File, error) {
	q := `INSERT INTO tutor_discussion_files (owner_id, file_path) VALUES ($1, $2)`
	_, err := repo.db.Query(q, item.OwnerID, item.FilePath)
	if err != nil {
		return nil, ErrFileCreation
	}
	return item, nil
}

func (repo *discussionRepo) GetParticipants(discussionId int64) ([]*Participant, error) {
	items := []*Participant{}
	q := `SELECT 
				users.id,
				participants.discussion_id,
				users.first_name,
				users.last_name,
				users.photo
			FROM tutor_discussion_participants AS participants
				LEFT JOIN tutor_users AS users
				ON users.id = participants.user_id 
					WHERE participants.discussion_id = $1`
	rows, err := repo.db.Query(q, discussionId)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		item := &participantRow{}
		err := rows.Scan(
			&item.ID,
			&item.DiscussionID,
			&item.FirstName,
			&item.LastName,
			&item.Photo,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item.toParticipant())
	}
	return items, nil
}

func (repo *discussionRepo) AddParticipant(participant *Participant) error {
	var unreadMessagesCnt int64
	q := `SELECT count(*) FROM tutor_discussion_messages 
			WHERE discussion_id = $1 AND sender_id != $2 AND is_read = false`
	err := repo.db.QueryRow(q, participant.DiscussionID, participant.ID).Scan(&unreadMessagesCnt)
	if err != nil {
		return err
	}

	ret, err := repo.db.Exec(
		`INSERT INTO tutor_discussion_participants (discussion_id, user_id, unread_messages_cnt) 
					VALUES ($1, $2, $3)
					ON CONFLICT DO NOTHING`,
		participant.DiscussionID,
		participant.ID,
		unreadMessagesCnt,
	)
	if err != nil {
		return err
	}
	n, err := ret.RowsAffected()
	if err != nil {
		return err
	}
	if n <= 0 {
		return ErrParticipantExists
	}
	return nil
}

func (repo *discussionRepo) RemoveParticipant(participant *Participant) error {
	ret, err := repo.db.Exec(
		`DELETE FROM tutor_discussion_participants WHERE discussion_id = $1 AND user_id = $2`,
		participant.DiscussionID,
		participant.ID,
	)
	if err != nil {
		return err
	}
	n, err := ret.RowsAffected()
	if err != nil {
		return err
	}
	if n <= 0 {
		return ErrParticipantNotFound
	}
	return nil
}

func (repo *discussionRepo) IncrementUnreadMessagesCnt(participant *Participant) error {
	q := `UPDATE tutor_discussion_participants SET unread_messages_cnt = unread_messages_cnt + 1
			WHERE discussion_id = $1 AND user_id = $2`
	ret, err := repo.db.Exec(q, participant.DiscussionID, participant.ID)
	if err != nil {
		return err
	}
	n, err := ret.RowsAffected()
	if err != nil {
		return err
	}
	if n <= 0 {
		return ErrParticipantUpdate
	}
	return nil
}

func (repo *discussionRepo) DecrementUnreadMessagesCnt(participant *Participant) error {
	q := `UPDATE tutor_discussion_participants SET unread_messages_cnt = unread_messages_cnt - 1
			WHERE discussion_id = $1 AND user_id = $2`
	ret, err := repo.db.Exec(q, participant.DiscussionID, participant.ID)
	if err != nil {
		return err
	}
	n, err := ret.RowsAffected()
	if err != nil {
		return err
	}
	if n <= 0 {
		return ErrParticipantUpdate
	}
	return nil
}
