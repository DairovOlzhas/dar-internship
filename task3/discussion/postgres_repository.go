package discussion

import (
	"database/sql"
	"fmt"
	"git.dar.tech/dareco-go/http"
	"git.dar.tech/dareco-go/utils/postgres"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

type discussionRaw struct {
	ID                 int64          `json:"id,omitempty"`
	FromId             sql.NullString `json:"from_id,omitempty"`
	ToId               sql.NullString `json:"to_id,omitempty"`
	CourseId           int64          `json:"course_id,omitempty"`
	IsActive           bool           `json:"is_active:omitempty"`
	UnreadMessagesCnt  int64          `json:"unread_messages_cnt:omitempty"`
	RecipientFirstName sql.NullString `json:"recipient_first_name,omitempty"`
	RecipientLastName  sql.NullString `json:"recipient_last_name,omitempty"`
	CourseName         sql.NullString `json:"course_name,omitempty"`
	Time               time.Time      `json:"time,omitempty"`
}

func (item *discussionRaw) toDiscussion() *Discussion {
	return &Discussion{
		ID:                 item.ID,
		FromId:             item.FromId.String,
		ToId:               item.ToId.String,
		CourseId:           item.CourseId,
		IsActive:           item.IsActive,
		UnreadMessagesCnt:  item.UnreadMessagesCnt,
		RecipientFirstName: item.RecipientFirstName.String,
		RecipientLastName:  item.RecipientLastName.String,
		CourseName:         item.CourseName.String,
		Time:               item.Time,
	}
}

type messageRaw struct {
	ID         int64          `json:"id,omitempty"`
	FromId     sql.NullString `json:"from_id,omitempty"`
	ToId       sql.NullString `json:"to_id,omitempty"`
	CourseId   int64          `json:"course_id,omitempty"`
	Text       sql.NullString `json:"text,omitempty"`
	FilePath   sql.NullString `json:"file_path,omitempty"`
	IsRead     bool           `json:"is_read,omitempty"`
	SentTime   time.Time      `json:"sent_time,omitempty"`
	SenderName sql.NullString `json:"sender_name,omitempty"`
}

func (item *messageRaw) toMessage() *Message {
	return &Message{
		ID:         item.ID,
		FromId:     item.FromId.String,
		ToId:       item.ToId.String,
		CourseId:   item.CourseId,
		Text:       item.Text.String,
		FilePath:   item.FilePath.String,
		IsRead:     item.IsRead,
		SentTime:   item.SentTime,
		SenderName: item.SenderName.String,
	}
}

type discussionRepo struct {
	db                   *sql.DB
	discussionsTableName string
	messagesTableName    string
	violationsTable      string
}

func NewPostgresRepoWithDB(db *sql.DB) (Repository, error) {
	var discussionsTableName = "tutor_discussions"
	var messagesTableName = "tutor_discussion_messages"
	var violationsTable = "tutor_discussion_violations"
	var messageRepoQueries = []string{
		`CREATE TABLE IF NOT EXISTS ` + discussionsTableName + ` (
			id serial PRIMARY KEY,
			from_id VARCHAR(255),
			to_id VARCHAR(255),
			course_id INTEGER NOT NULL default 0,
			is_active BOOLEAN default true,
			unread_messages_cnt INTEGER NOT NULL default 0,
			time TIMESTAMPTZ default now(),
			UNIQUE (from_id, to_id, course_id)
		)`,
		`CREATE TABLE IF NOT EXISTS ` + messagesTableName + ` (
			id serial PRIMARY KEY,
			from_id VARCHAR(255) NOT NULL,
			to_id VARCHAR(255),
			course_id INTEGER,
			text TEXT,
			file_path TEXT,
			is_read BOOLEAN NOT NULL default false,
			sent_time TIMESTAMPTZ default now()
		);`,
		//`CREATE TABLE IF NOT EXISTS ` + violationsTable + ` (
		//	id serial PRIMARY KEY,
		//	sender_id VARCHAR(255) NOT NULL,
		//	discussion_id INTEGER NOT NULL,
		//	text TEXT NOT NULL
		//);`,
		`CREATE TABLE IF NOT EXISTS tutor_discussion_files (
			id serial PRIMARY KEY,
			owner_id VARCHAR(255) NOT NULL,
			file_path VARCHAR(255) NOT NULL
		);`,
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
			messagesTableName,
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

func (repo *discussionRepo) FindAll(query FindQuery, params *http.PaginationParams) ([]*Discussion, error) {
	var items []*Discussion
	q := `SELECT
			discussions.id,
			discussions.from_id AS from_id,
			discussions.to_id AS to_id,
			discussions.course_id AS course_id,
			discussions.is_active AS is_active,
			discussions.unread_messages_cnt,
			users.first_name AS recipient_first_name,
			users.last_name AS recipient_last_name, 
			courses.name AS course_name,
			discussions.time AS time
		FROM ` + repo.discussionsTableName + ` AS discussions
			LEFT JOIN tutor_courses AS courses
			ON courses.id = discussions.course_id
			LEFT JOIN tutor_users AS users
			ON users.id = discussions.to_id`

	parts := []string{}
	values := []interface{}{}
	cnt := 0

	if query.FromId != nil {
		cnt++
		parts = append(parts, "from_id = $"+strconv.Itoa(cnt))
		values = append(values, *query.FromId)
	}
	if query.ToId != nil {
		cnt++
		parts = append(parts, "to_id = $"+strconv.Itoa(cnt))
		values = append(values, *query.ToId)
	}
	if query.CourseId != nil {
		cnt++
		parts = append(parts, "course_id = $"+strconv.Itoa(cnt))
		values = append(values, *query.CourseId)
	}
	if query.IsActive != nil {
		cnt++
		parts = append(parts, "is_active = $"+strconv.Itoa(cnt))
		values = append(values, *query.IsActive)
	}
	if query.Time != nil {
		cnt++
		parts = append(parts, "time <= $"+strconv.Itoa(cnt))
		values = append(values, *query.Time)
	}
	if query.RecipientFirstName != nil {
		cnt++
		parts = append(parts, "recipient_first_name = $"+strconv.Itoa(cnt))
		values = append(values, *query.RecipientFirstName)
	}
	if query.RecipientLastName != nil {
		cnt++
		parts = append(parts, "recipient_last_name = $"+strconv.Itoa(cnt))
		values = append(values, *query.RecipientLastName)
	}
	if query.CourseName != nil {
		cnt++
		parts = append(parts, "course_name = $"+strconv.Itoa(cnt))
		values = append(values, *query.CourseName)
	}

	if len(values) > 0 {
		q = q + " WHERE "
	}
	q = q + strings.Join(parts, " AND ")
	rows, err := repo.db.Query(q, values...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		item := &discussionRaw{}
		err := rows.Scan(
			&item.ID,
			&item.FromId,
			&item.ToId,
			&item.CourseId,
			&item.IsActive,
			&item.UnreadMessagesCnt,
			&item.RecipientFirstName,
			&item.RecipientLastName,
			&item.CourseName,
			&item.Time,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item.toDiscussion())
	}
	return items, nil
}

func (repo *discussionRepo) FindByID(id int64) (*Discussion, error) {
	item := &discussionRaw{}
	err := repo.db.QueryRow(
		`SELECT
			discussions.id,
			discussions.from_id,
			discussions.to_id,
			discussions.course_id,
			discussions.is_active,
			discussions.unread_messages_cnt,
			users.first_name,
			users.last_name,
			courses.name,
			discussions.time
		FROM `+repo.discussionsTableName+` AS discussions
			LEFT JOIN tutor_courses AS courses
			ON courses.id = discussions.course_id
			LEFT JOIN tutor_users AS users
			ON users.id = discussions.to_id 
		WHERE discussions.id = $1`,
		id,
	).Scan(
		&item.ID,
		&item.FromId,
		&item.ToId,
		&item.CourseId,
		&item.IsActive,
		&item.UnreadMessagesCnt,
		&item.RecipientFirstName,
		&item.RecipientLastName,
		&item.CourseName,
		&item.Time,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return item.toDiscussion(), nil
}

func (repo *discussionRepo) Create(item *Discussion) error {
	q := `INSERT INTO ` + repo.discussionsTableName +
		` (from_id, to_id, course_id, is_active, time)
			VALUES ($1,$2,$3,default,default)
			ON CONFLICT (from_id, to_id, course_id)
			DO UPDATE SET from_id=excluded.from_id
		RETURNING id`
	var id int64
	err := repo.db.QueryRow(q,
		item.FromId,
		item.ToId,
		item.CourseId).Scan(
		&id,
	)
	if err != nil {
		return ErrDiscussionCreation
	}
	item.ID = id
	return nil
}

func (repo *discussionRepo) Update(id int64, upd *Update) error {
	if id == 0 {
		return ErrUpdate
	}
	q := fmt.Sprintf("UPDATE %s SET", repo.discussionsTableName)

	parts := []string{}
	values := []interface{}{}
	cnt := 0
	if upd.UnreadMessagesCnt != nil {
		cnt++
		parts = append(parts, "unread_messages_cnt = $"+strconv.Itoa(cnt))
		values = append(values, *upd.UnreadMessagesCnt)
	}
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

func (repo *discussionRepo) FindAndUpdate(query FindQuery, upd *Update) error {
	q := fmt.Sprintf("UPDATE %s SET ", repo.discussionsTableName)

	var parts []string
	var values []interface{}
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

	if len(parts) <= 0 {
		return ErrNothingToUpdate
	}
	q = q + strings.Join(parts, " , ")

	parts = []string{}

	if *query.CourseId != 0 {
		cnt++
		parts = append(parts, "course_id = $"+strconv.Itoa(cnt))
		values = append(values, *query.CourseId)
	} else if *query.ToId != "" && *query.FromId != "" {
		cnt++
		parts = append(parts, fmt.Sprintf(
			"((from_id = $%d AND to_id = $%d) OR (from_id = $%d AND to_id = $%d))",
			cnt, cnt+1, cnt+1, cnt))
		cnt++
		values = append(values, *query.FromId, *query.ToId)
	} else {
		return ErrInvalidQuery
	}
	q = q + ` WHERE ` + strings.Join(parts, " AND ")

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

func (repo *discussionRepo) CreateMessage(item *Message) (*Message, error) {
	q := `
		INSERT INTO ` + repo.messagesTableName + `
			(from_id, to_id, course_id, text, file_path, is_read, sent_time) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	var id int64
	err := repo.db.QueryRow(q,
		item.FromId,
		item.ToId,
		item.CourseId,
		item.Text,
		item.FilePath,
		item.IsRead,
		item.SentTime).Scan(&id)
	if err != nil {
		return nil, ErrMessageCreation
	}
	item.ID = id
	return item, nil
}

func (repo *discussionRepo) FindMessageByID(id int64) (*Message, error) {
	item := &messageRaw{}
	err := repo.db.QueryRow(
		`SELECT
			messages.id,
			messages.from_id AS from_id,
			messages.to_id AS to_id,
			messages.course_id AS course_id,
			messages.text,
			messages.file_path,
			messages.is_read,
			messages.sent_time,
			users.first_name
		FROM `+repo.messagesTableName+` AS messages
			LEFT JOIN tutor_users AS users
			ON messages.from_id = users.id
		WHERE messages.id = $1`,
		id,
	).Scan(
		&item.ID,
		&item.FromId,
		&item.ToId,
		&item.CourseId,
		&item.Text,
		&item.FilePath,
		&item.IsRead,
		&item.SentTime,
		&item.SenderName,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return item.toMessage(), nil
}

func (repo *discussionRepo) UpdateMessage(id int64, upd *UpdateMessage) error {
	if id == 0 {
		return ErrUpdate
	}
	q := fmt.Sprintf("UPDATE %s SET", repo.messagesTableName)

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
		return ErrNotFound
	}
	return nil
}

func (repo *discussionRepo) FindAllMessages(query FindQuery) ([]*Message, error) {
	var items []*Message
	q := `SELECT
			messages.id,
			messages.from_id AS from_id,
			messages.to_id AS to_id,
			messages.course_id AS course_id,
			messages.text,
			messages.file_path,
			messages.is_read,
			messages.sent_time,
			users.first_name
		FROM ` + repo.messagesTableName + ` AS messages
			LEFT JOIN tutor_users AS users
			ON messages.from_id = users.id
		WHERE `
	cnt := 0
	values := []interface{}{}
	if *query.CourseId != 0 {
		cnt++
		q = q + `course_id = $1`
		values = append(values, query.CourseId)
	} else if *query.FromId != "" && *query.ToId != "" {
		cnt+=2
		q = q + `(from_id = $1 AND to_id = $2) OR (from_id = $2 AND to_id = $1)`
		values = append(values, *query.FromId, *query.ToId)
	} else {
		return nil, ErrInvalidQuery
	}
	if query.NotFromId != nil {
		cnt++
		q = q + ` AND from_id != $`+strconv.Itoa(cnt)
		values = append(values, *query.NotFromId)
	}
	if query.IsRead != nil {
		cnt++
		q = q + ` AND is_read != $`+strconv.Itoa(cnt)
		values = append(values, *query.IsRead)
	}
	q = q + ` ORDER BY sent_time ASC `
	rows, err := repo.db.Query(q, values...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		item := messageRaw{}
		err := rows.Scan(
			&item.ID,
			&item.FromId,
			&item.ToId,
			&item.CourseId,
			&item.Text,
			&item.FilePath,
			&item.IsRead,
			&item.SentTime,
			&item.SenderName,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item.toMessage())
	}
	return items, nil
}

func (repo *discussionRepo) DeleteMessages(fromId string, toId string, courseId int64) error {
	q := `DELETE FROM ` + repo.messagesTableName + ` WHERE `
	values := []interface{}{}
	if courseId != 0 {
		q = q + `course_id = $1`
		values = append(values, courseId)
	} else if fromId != "" && toId != "" {
		q = q + `(from_id = $1 AND to_id = $2) OR (from_id = $2 AND to_id = $1)`
		values = append(values, fromId, toId)
	} else {
		return ErrInvalidQuery
	}
	_, err := repo.db.Query(q, values...)
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

func (repo *discussionRepo) ExecQuery(q string, values ...interface{}) error {
	stmt, err := repo.db.Prepare(q)
	if err != nil {
		return err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Warn(err)
		}
	}()
	_, err = stmt.Exec(values...)
	if err != nil {
		return err
	}
	return nil
}
