package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// ------------ Константы, переменные ----------------------------------------------------
const (
	DEBUG LogType_t = 0
	INFO  LogType_t = 1
	ERROR LogType_t = 2
)

var (
	LogTypeMap = map[LogType_t]string{
		0: "DEBUG",
		1: "INFO",
		2: "ERROR"}
)

// ------------ Типы ----------------------------------------------------
type Logger_t struct {
	DebugMode      bool //режим отладки
	dbLogger       []dbLogger_t
	openedFiles    []*os.File  //файлы, открытые через AddFile()
	ioWriterLogger []io.Writer //интерфейс записи логов в io.Writer
}
type DBLogger_i interface {
	LogWrite(table string, logType byte, msg string, vars map[string]interface{}) error
}

type dbLogger_t struct {
	db    DBLogger_i //интерфейсы записи логов в бд
	table string     //название таблицы логов
}

type LogType_t byte

// ------------ Функции ----------------------------------------------------
func New() *Logger_t {
	l := new(Logger_t)
	return l
}
func (l *Logger_t) Close() {
	for _, file := range l.openedFiles { //закрываем все открытые файлы
		file.Close()
	}
	l = nil
}
func (l *Logger_t) AddDB(db DBLogger_i, table string) error {
	for _, v := range l.dbLogger {
		if v.db == db && v.table == table {
			return fmt.Errorf("бд уже добавлена")
		}
	}
	if db == nil {
		return fmt.Errorf("db = nil")
	}
	l.dbLogger = append(l.dbLogger, dbLogger_t{db, table})
	return nil
}
func (l *Logger_t) AddFile(path string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	l.openedFiles = append(l.openedFiles, f)
	l.ioWriterLogger = append(l.ioWriterLogger, f)
	return nil
}
func (l *Logger_t) AddIOWriter(w io.Writer) error {
	if w == nil {
		return fmt.Errorf("io.Writer = nil")
	}
	l.ioWriterLogger = append(l.ioWriterLogger, w)
	return nil
}

// Запись лога в приёмники
// logType - тип лога, msg - текст сообщения, vars - переменные, преобразуются в JSON строку
func (l *Logger_t) Print(logType LogType_t, msg string, vars map[string]interface{}) {
	//пропускаем DEBUG логи, если режим отладки отключен
	if (logType == DEBUG) && !l.DebugMode {
		return
	}
	err := l.printIntoDB(logType, msg, vars) //запись логов в бд
	if err != nil {
		l.printIntoIOWriters(ERROR, fmt.Sprintf("logger: printIntoDB(): %s", err), nil) //при ошибке пишем лог в ioWriterLogger
	}
	l.printIntoIOWriters(logType, msg, vars) //запись логов в ioWriterLogger
}

// записывает лог в БД.
func (l *Logger_t) printIntoDB(logType LogType_t, msg string, vars map[string]interface{}) error {
	for _, v := range l.dbLogger {
		err := v.db.LogWrite(v.table, byte(logType), msg, vars)
		if err != nil {
			return err
		}
	}
	return nil
}

// записывает лог в ioWriter
func (l *Logger_t) printIntoIOWriters(logType LogType_t, msg string, vars map[string]interface{}) {
	miltiWriter := io.MultiWriter(l.ioWriterLogger...)
	str := fmt.Sprintf("%s  %s\t", time.Now().Format("2006-01-02 15:04:05"), LogTypeMap[logType])
	if msg != "" {
		str += fmt.Sprintf("%s ", msg)
	}
	varsJSON, err := json.Marshal(vars)
	if (err == nil) && (string(varsJSON) != "null") {
		str += string(varsJSON)
	}
	str += "\n"
	miltiWriter.Write([]byte(str))
}
