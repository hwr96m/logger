package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	dbcon "github.com/hwr96m/db-connector"
	"github.com/lib/pq"
)

// ------------ Структуры -----------------------------------------------------------------
type LoggerPsql struct {
	*dbcon.DB_t
}

//------------ Функции -------------------------------------------------------------------

// создаём экземпляр LoggerPsql, реализующий интерфейс DBLogger_i
func NewPsql(db *dbcon.DB_t) *LoggerPsql {
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
