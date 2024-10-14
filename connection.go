package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// establish connection to the database
type ConnectionExecutor interface {
	Connect() error
	Close() error
	GetDB() *sql.DB
}

// concrete implementation of ConnectionExecutor
type postgresConnection struct {
	db     *sql.DB // db instance handle
	Host   string  `json:"host"`
	DbName string  `json:"dbname"`
	Port   string  `json:"port"`
	User   string  `json:"user"`
	Pass   string  `json:"pass"`
	Ssl    string  `json:"ssl"`
}

// connect to the postgres db
func (c *postgresConnection) Connect() error {
	if c.Host == "" {
		return fmt.Errorf("host not set")
	} else if c.DbName == "" {
		return fmt.Errorf("database not set")
	} else if c.Port == "" {
		return fmt.Errorf("port not set")
	} else if c.User == "" {
		return fmt.Errorf("user not set")
	} else if c.Pass == "" {
		return fmt.Errorf("creds not set")
	} else if c.Ssl == "" {
		return fmt.Errorf("ssl mode not set")
	}

	dsn := fmt.Sprintf("host=%s dbname=%s user=%s password=%s port=%s sslmode=%s", c.Host, c.DbName, c.User, c.Pass, c.Port, c.Ssl)

	// open a connection to postgres
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to db => %s", err.Error())
	}

	// verify the connection
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping db => %s", err.Error())
	}

	c.db = db
	return nil
}

// close the connection to the postgres db
func (c *postgresConnection) Close() error {
	return c.db.Close()
}

// get a handle to the db instance
func (c *postgresConnection) GetDB() *sql.DB {
	return c.db
}

// constructor for orm
func NewConnection(host, dbname, user, pass, port, ssl string) ConnectionExecutor {
	return &postgresConnection{
		Host:   host,
		DbName: dbname,
		User:   user,
		Pass:   pass,
		Port:   port,
		Ssl:    ssl,
	}
}
