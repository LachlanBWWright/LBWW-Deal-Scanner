package discord

import "strings"

const errorLogQueueSize = 100

type errorSender interface {
	SendError(channelId, message string) error
}

// ErrorLogWriter forwards warning and error log lines to Discord without
// blocking application logging. Delivery failures are intentionally ignored
// to prevent a Discord outage from creating a recursive logging loop.
type ErrorLogWriter struct {
	channelID string
	sender    errorSender
	messages  chan string
}

func NewErrorLogWriter(channelID string, sender errorSender) *ErrorLogWriter {
	writer := &ErrorLogWriter{
		channelID: channelID,
		sender:    sender,
		messages:  make(chan string, errorLogQueueSize),
	}
	go writer.run()
	return writer
}

func (w *ErrorLogWriter) Write(data []byte) (int, error) {
	line := strings.TrimSpace(string(data))
	if w.channelID == "" || !isErrorLogLine(line) {
		return len(data), nil
	}
	select {
	case w.messages <- line:
	default:
	}
	return len(data), nil
}

func (w *ErrorLogWriter) run() {
	for message := range w.messages {
		message = strings.ReplaceAll(message, "@", "@\u200b")
		if len(message) > 1900 {
			message = message[:1900] + "..."
		}
		_ = w.sender.SendError(w.channelID, "Warning: "+message)
	}
}

func isErrorLogLine(line string) bool {
	lower := strings.ToLower(line)
	return strings.Contains(lower, "error") ||
		strings.Contains(lower, "fail") ||
		strings.Contains(lower, "warn") ||
		strings.Contains(lower, "fatal") ||
		strings.Contains(lower, "panic")
}
