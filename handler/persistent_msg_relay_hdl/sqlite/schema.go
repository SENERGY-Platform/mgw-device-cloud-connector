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
	PRIMARY KEY (id),
    INDEX (day_hour)
    INDEX (msg_num)
    INDEX (day_hour, msg_num)
);
`
