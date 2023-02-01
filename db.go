package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	dbstruct "home.rep/go-libs/db-struct.git"
)

// ------------ Структуры -----------------------------------------------------------------
type LoggerPsql struct {
	*dbstruct.DB_t
}

//------------ Функции -------------------------------------------------------------------

// -_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_-_
// подключение к базе
/*func OpenPostgres(conf *dbstruct.Config_t) (db *LoggerPsql, err error) {
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
	db = new(LoggerPsql)
	db.Timeout = 30 * time.Second
	db.Config = conf
	db.DB, err = sql.Open("postgres",
		fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			conf.Host,
			conf.Port,
			conf.Login,
			conf.Password,
			conf.Database))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("func OpenPostgres: db.PingContext: %w", err)
	}
	return db, err
}*/

// возвращает структуру, реализующую интерфейс DBLogger_i
func GetPsql(db *dbstruct.DB_t) *LoggerPsql {
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
