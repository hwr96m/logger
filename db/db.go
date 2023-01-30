package logdb

// ------------ Структуры -----------------------------------------------------------------

type LogDB_i interface {
	LogWrite(table, logType, msg string, vars map[string]interface{}) error
	Close() (err error)
}
