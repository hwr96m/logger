package logger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// ------------ Структуры -----------------------------------------------------------------
type LoggerPsql struct {
	*DB_t
}

type Config_t struct {
	Host     string `json:"Host"`
	Port     string `json:"Port"`
	Database string `json:"Database"`
	Login    string `json:"Login"`
	Password string `json:"Password"`
	Scheme   string `json:"Scheme"`
}

type DB_t struct {
	*sql.DB
	Config  *Config_t
	Timeout time.Duration
}

//------------ Функции -------------------------------------------------------------------

// -_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_

// подключение к базе
func OpenPostgres(conf *Config_t) (db *DB_t, err error) {
	if conf.Login == "" {
		return nil, fmt.Errorf("parameter Login is null")
	}
	if conf.Password == "" {
		return nil, fmt.Errorf("parameter Password is null")
	}
	if conf.Database == "" {
		return nil, fmt.Errorf("parameter Dbname is null")
	}
	if conf.Host == "" {
		conf.Host = "localhost"
	}
	if conf.Port == "" {
		conf.Port = "5432"
	}
	db = new(DB_t)
	db.Timeout = 30 * time.Second
	db.Config = conf
	db.DB, err = sql.Open("postgres",
		fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			db.Config.Host,
			db.Config.Port,
			db.Config.Login,
			db.Config.Password,
			db.Config.Database))
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("func OpenPostgres(): %w", err)
	}
	return db, err
}

// возвращает структуру, реализующую интерфейс DBLogger_i
func GetPsql(db *DB_t) *LoggerPsql {
	psql := new(LoggerPsql)
	psql.DB_t = db
	return psql
}

//------------ Logger -----------------------------------------------------------------

func (db *LoggerPsql) LogWrite(table string, logType byte, msg string, vars map[string]interface{}) error {
	var (
		err      error
		varsJSON []byte
	)
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	varsJSON, err = json.Marshal(vars)
	if err != nil {
		return fmt.Errorf("func LogWrite: json.Marshal: %w", err)
	}
	qry := fmt.Sprintf(`INSERT INTO %s.%s.%s (time,type,msg, vars)	VALUES ($1,$2,$3,$4);`, pq.QuoteIdentifier(db.Config.Database), pq.QuoteIdentifier(db.Config.Scheme), pq.QuoteIdentifier(table))
	//fmt.Println(qry)
	_, err = db.ExecContext(ctx, qry, time.Now(), logType, msg, varsJSON)
	if err != nil {
		return fmt.Errorf("func LogWrite: db.ExecContext: %s", err)
	}
	return nil
}
