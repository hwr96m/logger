package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

//------------ Константы, переменные ----------------------------------------------------
const (
	DEBUG LogType_t = "DEBUG"
	INFO  LogType_t = "INFO"
	ERROR LogType_t = "ERROR"
)

var ()

//------------ Типы ----------------------------------------------------
type Logger_t struct {
	DebugMode      bool         //режим отладки
	dbLogger       []DBLogger_t //интерфейсы записи логов в бд
	openedFiles    []*os.File   //файлы, открытые через AddFile()
	ioWriterLogger []io.Writer  //интерфейс записи логов в io.Writer
}
type DBLogger_t interface {
	Write(t time.Time, logType, msg string, vars map[string]interface{}) error
}
type LogType_t string

//------------ Функции ----------------------------------------------------
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
func (l *Logger_t) AddDB(db DBLogger_t) error {
	for _, v := range l.dbLogger {
		if v == db {
			return fmt.Errorf("бд уже добавлена")
		}
	}
	if db == nil {
		return fmt.Errorf("db = nil")
	}
	l.dbLogger = append(l.dbLogger, db)
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

//Запись лога в приёмники
//logType - тип лога, msg - текст сообщения, vars - переменные, преобразуются в JSON строку
func (l *Logger_t) Print(logType LogType_t, msg string, vars map[string]interface{}) {
	//пропускаем DEBUG логи, если режим отладки отключен
	if (logType == DEBUG) && !l.DebugMode {
		return
	}
	err := l.printIntoDB(logType, msg, vars) //запись логов в бд
	if err != nil {
		l.printIntoIOWriters(string(ERROR), fmt.Sprintf("logger: printIntoDB(): %s", err), nil) //при ошибке пишем лог в ioWriterLogger
	}
	l.printIntoIOWriters(string(logType), msg, vars) //запись логов в ioWriterLogger
}

//записывает лог в БД.
func (l *Logger_t) printIntoDB(logType LogType_t, msg string, vars map[string]interface{}) error {
	for i := range l.dbLogger {
		err := l.dbLogger[i].Write(time.Now(), string(logType), msg, vars)
		if err != nil {
			return err
		}
	}
	return nil
}

//записывает лог в ioWriter
func (l *Logger_t) printIntoIOWriters(logType, msg string, vars map[string]interface{}) {
	miltiWriter := io.MultiWriter(l.ioWriterLogger...)
	str := fmt.Sprintf("%s  %s\t", time.Now().Format("2006-01-02 15:04:05"), logType)
	if msg != "" {
		str += fmt.Sprintf("%s. ", msg)
	}
	varsJSON, err := json.Marshal(vars)
	if (err == nil) && (string(varsJSON) != "null") {
		str += string(varsJSON)
	}
	str += "\n"
	miltiWriter.Write([]byte(str))
}
