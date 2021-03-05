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
	CreatedAt time.Time `pg:"default:now(),notnull"`
	UpdatedAt time.Time
}

// BeforeUpdate injects time.Now in to updated_at
func (b *BaseModel) BeforeUpdate(ctx context.Context) (context.Context, error) {
	b.UpdatedAt = time.Now()
	return ctx, nil
}

// User is a User model
type User struct {
	BaseModel
	ID        int  `pg:",notnull,pk"`
	IsBot     bool `pg:"default:false,notnull"`
	FirstName string
	LastName  string
	Username  string
	Feeds     []*Feed `pg:"rel:has-many,join_fk:user_id"`
	LastFetch *time.Time
	tableName struct{} `pg:"users"`
}

// Feed is a User model
type Feed struct {
	BaseModel
	ID        string   `pg:"type:uuid,default:gen_random_uuid(),pk"`
	UserID    int      `pg:",notnull,on_delete:CASCADE,unique:unq_feeds_user_id_link"`
	IsRSS     bool     `pg:"default:true,notnull"`
	Link      string   `pg:",notnull,unique:unq_feeds_user_id_link"`
	Name      string   `pg:",notnull"`
	User      *User    `pg:"rel:has-one,fk:user_id"`
	tableName struct{} `pg:"feeds"`
	Rhash     string
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
		log.Fatal().Err(err).Msg("Error parsing DATABASE_URL")
	}
	db = pg.Connect(d)
	return

}

func createSchema(db *pg.DB) error {
	models := []interface{}{
		(*User)(nil),
		(*Feed)(nil),
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

func getUserFeeds(userID int) ([]*Feed, error) {
	var feeds []*Feed
	if err := db.Model(&feeds).Where("user_id = ?", userID).Select(); err != nil {
		return nil, err
	}
	return feeds, nil
}

func getUsers() ([]*User, error) {
	var users []*User
	if err := db.Model(&users).Relation("Feeds").Select(); err != nil {
		return nil, err
	}
	return users, nil
}

func addUser(message *TelegramMessage) (int, error) {
	user := User{
		ID:       message.From.UserID,
		IsBot:    message.From.IsBot,
		Username: message.From.Username,
	}
	_, err := db.Model(&user).OnConflict("(id) DO UPDATE").
		Set("is_bot = EXCLUDED.is_bot, username = EXCLUDED.username").
		Insert()
	if err != nil {
		log.Error().Err(err).
			Int("userID", message.From.UserID).
			Msg("Error inserting user")
		return 0, err
	}
	return user.ID, nil
}

func addFeed(rawFeed *RawFeed, message *TelegramMessage) bool {
	feed := Feed{
		UserID: message.From.UserID,
		IsRSS:  rawFeed.IsRSS,
		Link:   rawFeed.URL,
		Name:   rawFeed.Name,
	}
	if _, err := db.Model(&feed).OnConflict("(user_id,link) DO NOTHING").
		Insert(); err != nil {
		log.Error().Err(err).
			Int("userID", message.From.UserID).
			Str("feedURL", rawFeed.URL).Msg("Error saving new feed")
		return false
	}
	return true
}

func updateLastFetch(userID int) bool {
	_, err := db.Model((*User)(nil)).
		Set("last_fetch = ?", time.Now()).
		Where("id = ?", userID).
		Update()
	if err != nil {
		log.Error().Err(err).Int("userID", userID).Msg("Failed to update lastFetch")
		return false
	}
	return true
}
