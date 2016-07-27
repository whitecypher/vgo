package logging

type Logger interface{
	Debug(...interface{}) string
	Debugf(string, ...interface{}) string
	Error(...interface{}) string
	Errorf(string, ...interface{}) string
	Info(...interface{}) string
	Infof(string, ...interface{}) string
	Panic(...interface{})
	Panicf(string, ...interface{})
	Success(...interface{}) string
	Successf(string, ...interface{}) string
	Warn(...interface{}) string
	Warnf(string, ...interface{}) string
}