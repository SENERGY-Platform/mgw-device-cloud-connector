package sqlite

var messagesTable = `
CREATE TABLE IF NOT EXISTS messages
(
    id       TEXT NOT NULL,
    day_hour INTEGER NOT NULL,
    msg_num  INTEGER NOT NULL,
	topic    TEXT NOT NULL,
	payload  BLOB NULL,
	msg_time INTEGER NOT NULL,
	PRIMARY KEY (id)
);
`

var messagesIndexes = []string{
	"CREATE INDEX IF NOT EXISTS messages_day_hour ON messages (day_hour);",
	"CREATE INDEX IF NOT EXISTS messages_msg_num ON messages (msg_num);",
	"CREATE UNIQUE INDEX IF NOT EXISTS messages_day_hour_msg_num ON messages (day_hour, msg_num);",
}
