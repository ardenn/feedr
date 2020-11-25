package main

import (
	"context"
	"os"
	"time"

	"github.com/go-pg/pg/v10"
)

// BaseModel implements shared model fields
type BaseModel struct {
	ID        string    `pg:",pk;default:gen_random_uuid();type:uuid"`
	CreatedAt time.Time `pg:",notnull;default:now()"`
	UpdatedAt time.Time `pg:",notnull"`
}

// BeforeUpdate injects time.Now in to updated_at
func (b *BaseModel) BeforeUpdate(ctx context.Context) (context.Context, error) {
	b.UpdatedAt = time.Now()
	return ctx, nil
}

// PgUser is a User model
type PgUser struct {
	BaseModel
	UserID    int  `pg:",unique;,notnull"`
	ChatID    int  `pg:",unique;,notnull"`
	IsBot     bool `pg:",notnull,default:false"`
	FirstName string
	LastName  string
	Username  string
	Feeds     []*PgFeed `pg:"rel:has-many;join_fk:user_id"`
	tableName struct{}  `pg:"users"`
}

// PgFeed is a User model
type PgFeed struct {
	BaseModel
	UserID    string   `pg:"type:uuid"`
	IsRSS     bool     `pg:",notnull;default:true"`
	Link      string   `pg:",notnull"`
	Name      string   `pg:",notnull"`
	User      *PgUser  `pg:"rel:has-one;fk:user_id"`
	tableName struct{} `pg:"feeds"`
}

func dbConnect() *pg.DB {
	db := pg.Connect(&pg.Options{
		Addr:     os.Getenv("DB_HOST") + os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_NAME"),
	})
	return db
}
