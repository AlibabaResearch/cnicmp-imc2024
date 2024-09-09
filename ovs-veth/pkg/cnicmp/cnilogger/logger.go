package cnilogger

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Level = zapcore.Level
type Field = zap.Field

type CNILogger struct {
	logger         *zap.Logger
	level          *zap.AtomicLevel
	bufferedWriter *zapcore.BufferedWriteSyncer
}

type CNILoggerError struct {
	BufferError error
	LoggerError error
}

const (
	DebugLevel = zapcore.DebugLevel
	InfoLevel  = zapcore.InfoLevel
	WarnLevel  = zapcore.WarnLevel
	ErrorLevel = zapcore.ErrorLevel
	PanicLevel = zapcore.PanicLevel
	FatalLevel = zapcore.FatalLevel
)

func getTimeDigitStr(targetTime time.Time) string {
	year, month, day := targetTime.Date()
	hour, min, sec := targetTime.Clock()
	return fmt.Sprintf("%d%02d%02d%02d%02d%02d", year, month, day, hour, min, sec)
}

func initLogEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func initLogWriter(filenameTpl string, maxSplitSizeMB int, maxbufferSizeMB int, flushInterval time.Duration) *zapcore.BufferedWriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:  filenameTpl,
		MaxSize:   maxSplitSizeMB,
		Compress:  false,
		LocalTime: true,
	}
	return &zapcore.BufferedWriteSyncer{
		WS:            zapcore.AddSync(lumberJackLogger),
		Size:          maxbufferSizeMB * 1024 * 1024,
		FlushInterval: flushInterval,
	}
}

func NewCNILogger(level Level, filenameTpl string, maxSplitSizeMB int, maxbufferSizeMB int, flushInterval time.Duration) *CNILogger {
	atomicLevel := zap.NewAtomicLevelAt(level)
	writer := initLogWriter(filenameTpl, maxSplitSizeMB, maxbufferSizeMB, flushInterval)
	encoder := initLogEncoder()
	core := zapcore.NewCore(encoder, writer, atomicLevel)

	record_fields = make([]zap.Field, 2)
	record_fields[0] = zap.String("type", "start")
	record_fields[1] = zap.String("type", "end")

	return &CNILogger{logger: zap.New(core), level: &atomicLevel, bufferedWriter: writer}
}

func (l *CNILogger) SetLevel(level Level) {
	if l.level != nil {
		l.level.SetLevel(level)
	}
}

func (l *CNILogger) SyncOrStop(doSync bool) error {
	// Sync/Stop log buffer first and then sync log file
	var bufferError, loggerError error
	if doSync {
		bufferError = l.bufferedWriter.Sync()
	} else {
		bufferError = l.bufferedWriter.Stop()
	}
	loggerError = l.logger.Sync()
	return &CNILoggerError{BufferError: bufferError, LoggerError: loggerError}
}

func (l *CNILogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, fields...)
}

func (l *CNILogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, fields...)
}

func (l *CNILogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, fields...)
}

func (l *CNILogger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, fields...)
}

func (l *CNILogger) Panic(msg string, fields ...Field) {
	l.logger.Panic(msg, fields...)
}

func (l *CNILogger) Fatal(msg string, fields ...Field) {
	l.logger.Fatal(msg, fields...)
}

func (e *CNILoggerError) Error() string {
	return fmt.Sprintf("BufferError: %v, LoggerError: %v", e.BufferError, e.LoggerError)
}

var std = NewCNILogger(
	DebugLevel,
	"/home/cni/cnicmp_logs/cni_logs/ovs_cni/ovs.log",
	10,          // 10MB - max split file size
	1,           // 1MB  - max buffer size
	time.Minute, // 1min - max buffer time
)

var record_fields []zap.Field

func Default() *CNILogger         { return std }
func ReplaceDefault(l *CNILogger) { std = l }

func SetLevel(level Level) { std.SetLevel(level) }

func Debug(msg string, fields ...Field) { std.Debug(msg, fields...) }
func Info(msg string, fields ...Field)  { std.Info(msg, fields...) }
func Warn(msg string, fields ...Field)  { std.Warn(msg, fields...) }
func Error(msg string, fields ...Field) { std.Error(msg, fields...) }
func Panic(msg string, fields ...Field) { std.Panic(msg, fields...) }
func Fatal(msg string, fields ...Field) { std.Fatal(msg, fields...) }
func Sync() error                       { return std.SyncOrStop(true) }
func Stop() error                       { return std.SyncOrStop(false) }

func RecordStart() time.Time {
	return time.Now()
}

func RecordEnd(recordID string, subject string, start time.Time) {
	elapsed := time.Since(start)
	std.Info(subject, zap.String("record_id", recordID), zap.Time("start_t", start), zap.Int64("elapsed_ms", elapsed.Nanoseconds()/1000000))
}

func RecordEndWithMsg(recordID string, subject string, start time.Time, msg_name string, msg_body string) {
	elapsed := time.Since(start)
	std.Info(subject, zap.String("record_id", recordID), zap.Time("start_t", start), zap.Int64("elapsed_ms", elapsed.Nanoseconds()/1000000), zap.String(msg_name, msg_body))
}
