package main

import (
	"context"
	"os"
	"time"

	"github.com/go-pg/pg/extra/pgdebug"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/rs/zerolog/log"
)

// BaseModel implements shared model fields
type BaseModel struct {
	ID        string    `pg:"type:uuid,default:gen_random_uuid()"`
	CreatedAt time.Time `pg:"default:now(),notnull"`
	UpdatedAt time.Time
}

// BeforeUpdate injects time.Now in to updated_at
func (b *BaseModel) BeforeUpdate(ctx context.Context) (context.Context, error) {
	b.UpdatedAt = time.Now()
	return ctx, nil
}

// PgUser is a User model
type PgUser struct {
	BaseModel
	UserID    int  `pg:",unique,notnull"`
	ChatID    int  `pg:",unique,notnull"`
	IsBot     bool `pg:"default:false,notnull"`
	FirstName string
	LastName  string
	Username  string
	Feeds     []*PgFeed `pg:"rel:has-many,join_fk:user_id"`
	LastFetch time.Time
	tableName struct{} `pg:"users"`
}

// PgFeed is a User model
type PgFeed struct {
	BaseModel
	UserID    string   `pg:"type:uuid,notnull,on_delete:CASCADE"`
	IsRSS     bool     `pg:"default:true,notnull"`
	Link      string   `pg:",notnull"`
	Name      string   `pg:",notnull"`
	User      *PgUser  `pg:"rel:has-one,fk:user_id"`
	tableName struct{} `pg:"feeds"`
}

func dbConnect() (db *pg.DB) {
	if os.Getenv("APP_ENV") != "PROD" {
		db = pg.Connect(&pg.Options{
			Addr:     os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Database: os.Getenv("DB_NAME"),
		})
		return
	}
	d, err := pg.ParseURL(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal().Str("error", err.Error()).Msg("Error parsing DATABASE_URL")
	}
	db = pg.Connect(d)
	return

}

func createSchema(db *pg.DB) error {
	models := []interface{}{
		(*PgUser)(nil),
		(*PgFeed)(nil),
	}
	for _, model := range models {
		err := db.Model(model).CreateTable(&orm.CreateTableOptions{
			Temp:          false,
			IfNotExists:   true,
			FKConstraints: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func logQueries(*pg.DB) {
	debug := os.Getenv("APP_ENV") != "PROD"
	db.AddQueryHook(pgdebug.DebugHook{
		// Print all queries.
		Verbose: debug,
	})
}

func getUsers() ([]*PgUser, error) {
	var users []*PgUser
	if err := db.Model(&users).Relation("Feeds").Select(); err != nil {
		return nil, err
	}
	return users, nil
}

func addUser(message Message) string {
	user := PgUser{
		UserID:   message.From.UserID,
		IsBot:    message.From.IsBot,
		ChatID:   message.Chat.ChatID,
		Username: message.From.Username,
	}
	if _, err := db.Model(user).OnConflict("(user_id,chat_id) DO UPDATE").
		Set("is_bot = EXCLUDED.is_bot, username = EXCLUDED.username").
		Insert(); err != nil {
		log.Error().Str("error", err.Error()).Int("userID", message.From.UserID).Msg("Error inserting user")
	}
	return user.ID
}

func addFeed(userID int, fireFeed FireFeed, message Message) bool {
	var id string
	err := db.Model((*PgUser)(nil)).Column("id").
		Where("user_id = ?", userID).
		Limit(1).
		Select(&id)
	if err != nil {
		log.Error().Str("error", err.Error()).Int("userID", userID).
			Str("feedURL", fireFeed.URL).Msg("Error getting new feed user")
		return false
	}
	if id == "" {
		log.Info().Int("userID", userID).Str("feedURL", fireFeed.URL).Msg("User not found, attempting creation")
		id = addUser(message)
		if id == "" {
			return false
		}
	}
	feed := PgFeed{
		UserID: id,
		IsRSS:  fireFeed.IsRSS,
		Link:   fireFeed.URL,
		Name:   fireFeed.Name,
	}
	if _, err := db.Model(feed).Insert(); err != nil {
		log.Error().Str("error", err.Error()).Int("userID", userID).
			Str("feedURL", fireFeed.URL).Msg("Error saving new feed")
		return false
	}
	return true
}
