package sematextHook

import (
	"encoding/json"
	"net/url"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// a hook that sends messages to sematext
type sematextHook struct {
	client      *resty.Client
	baseUrl     string
	facility    string
	levelMapper func(level logrus.Level) string
	environment string
}

// Create a new sematextHook.
//
// - client: a configured resty.Client
// - baseUrl: the sematext url, something like https://logsene-receiver.sematext.com/<APP_TOKEN>/
// - group: logsene_type, most likely your products name, e.g. myservice
// - facility: the very name of the service, e.g. api
//
//noinspection GoUnusedExportedFunction
func NewSematextHook(client *resty.Client, baseUrl, group, facility, environment string) (*sematextHook, error) {
	s, e := url.Parse(baseUrl)
	if e != nil {
		return nil, e
	}

	groupPath, _ := url.Parse(group)

	basePath := s.ResolveReference(groupPath).String()

	return &sematextHook{
		client:      client,
		baseUrl:     basePath,
		facility:    facility,
		levelMapper: AsLogrusLevel,
		environment: environment}, nil
}

// set your own levelMapper
func (s *sematextHook) WithLevelMapper(levelMapper func(level logrus.Level) string) {
	s.levelMapper = levelMapper
}

func (s sematextHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}

func (s sematextHook) Fire(entry *logrus.Entry) error {
	return s.asyncFire(entry)
}

func (s sematextHook) asyncFire(entry *logrus.Entry) error {
	go func() {
		err := s.syncFire(entry)
		if err != nil {
			print("Failed to fire hook, got " + err.Error())
		}
	}()
	return nil
}

type SematextMessage struct {
	Severity     string    `json:"Severity,omitempty"`
	Time         time.Time `json:"Time,omitempty"`
	Environment  string    `json:"environment,omitempty"`
	Facility     string    `json:"facility,omitempty"`
	Host         string    `json:"host,omitempty"`
	Level        int       `json:"level,omitempty"`
	ShortMessage string    `json:"short_message,omitempty"`
	FullMessage  string    `json:"full_message,omitempty"`
}

// level mapper that produces CAPITAL_CASE level strings that look very much alike the ones by logback
// and probably others.
func AsLogbackLevel(level logrus.Level) string {
	switch level {
	case logrus.TraceLevel:
		return "TRACE"
	case logrus.DebugLevel:
		return "DEBUG"
	case logrus.InfoLevel:
		return "INFO"
	case logrus.WarnLevel:
		return "WARN"
	case logrus.ErrorLevel:
		return "ERROR"
	case logrus.FatalLevel:
		return "FATAL"
	case logrus.PanicLevel:
		return "PANIC"
	}

	return "UNKNOWN"
}

// level mapper that produces lower_case level strings that look very much alike the ones by logrus
// and probably others.
func AsLogrusLevel(level logrus.Level) string {
	switch level {
	case logrus.TraceLevel:
		return "trace"
	case logrus.DebugLevel:
		return "debug"
	case logrus.InfoLevel:
		return "info"
	case logrus.WarnLevel:
		return "warning"
	case logrus.ErrorLevel:
		return "error"
	case logrus.FatalLevel:
		return "fatal"
	case logrus.PanicLevel:
		return "panic"
	}

	return "unknown"
}

func (s sematextHook) syncFire(entry *logrus.Entry) error {
	hostname, _ := os.Hostname()
	message := &SematextMessage{
		Severity:     s.logLevelMapper(entry.Level),
		FullMessage:  entry.Message,
		ShortMessage: entry.Message[:min(len(entry.Message), 255)],
		Level:        int(entry.Level),
		Environment:  s.environment,
		Host:         hostname,
		Time:         entry.Time,
		Facility:     s.facility,
	}

	if entry.Data != nil {
		err := s.sendWithExtraData(message, entry.Data)
		if err == nil {
			return nil
		}
	}

	return s.sendLogMessage(message)
}

func (s sematextHook) sendLogMessage(message interface{}) error {
	_, e := s.client.R().SetBody(message).Post(s.baseUrl)
	return e
}

func (s sematextHook) sendWithExtraData(message *SematextMessage, fields logrus.Fields) error {
	messageJson, err := json.Marshal(message)
	if err != nil {
		return err
	}
	var data map[string]interface{}

	err = json.Unmarshal(messageJson, &data)
	if err != nil {
		return err
	}

	for k, v := range fields {
		if data[k] != nil {
			// do not allow overwriting
			continue
		}
		str, ok := v.(string)
		if ok && str == "" {
			// skip empty fields
			continue
		}
		data[k] = v
	}

	return s.sendLogMessage(data)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s sematextHook) logLevelMapper(level logrus.Level) string {
	return AsLogbackLevel(level)
}
