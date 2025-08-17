PRAGMA foreign_keys = 1;

CREATE TABLE USERS (
	user_id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_email TEXT NOT NULL UNIQUE,
	user_name TEXT NOT NULL,
	user_hash TEXT NOT NULL,
	user_description TEXT,
	user_isadmin BOOLEAN DEFAULT 0,
	user_isbot BOOLEAN DEFAULT 0,
	user_ishidden BOOLEAN DEFAULT 0,
	user_isdeactivated BOOLEAN DEFAULT 0,
	user_mustchangename BOOLEAN DEFAULT 0,
	user_mustchangepassword BOOLEAN DEFAULT 0,
	user_invited DATETIME,
	user_joined DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE CHALLENGES (
	challenge_id INTEGER PRIMARY KEY AUTOINCREMENT,
	challenge_title TEXT NOT NULL,
	challenge_tags TEXT,
	challenge_description TEXT,
	challenge_flags TEXT NOT NULL,
	challenge_points INTEGER DEFAULT 1,
	challenge_featured BOOLEAN DEFAULT 0,
	challenge_hidden BOOLEAN DEFAULT 0,
	challenge_uploaded DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE SOLVES (
	solve_id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	challenge_id INTEGER NOT NULL,
	solve_datetime DATETIME DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(user_id, challenge_id),
	FOREIGN KEY (user_id) REFERENCES USERS(user_id),
	FOREIGN KEY (challenge_id) REFERENCES CHALLENGES(challenge_id)
);

CREATE TABLE NOTICES (
	notice_id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	notice_title TEXT NOT NULL,
	notice_content TEXT NOT NULL,
	notice_posted DATETIME DEFAULT CURRENT_TIMESTAMP,
	notice_expiry DATETIME,
	FOREIGN KEY (user_id) REFERENCES USERS(user_id)
);

CREATE TABLE NOTICES_QUEUE (
	notice_queue_id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	notice_queue_title TEXT NOT NULL,
	notice_queue_content TEXT NOT NULL,
	notice_queue_scheduled DATETIME,
	notice_queue_expiry DATETIME,
	notice_queue_crosspost_to_discord BOOLEAN DEFAULT 0,
	FOREIGN KEY (user_id) REFERENCES USERS(user_id)
);

CREATE TABLE MESSAGES (
	message_id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	message_content TEXT NOT NULL,
	message_datetime DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES USERS(user_id)
);

CREATE TABLE MESSAGES_QUEUE (
	message_queue_id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	message_queue_content TEXT NOT NULL,
	message_queue_scheduled_datetime DATETIME,
	message_queue_crosspost_to_discord BOOLEAN DEFAULT 0,
	FOREIGN KEY (user_id) REFERENCES USERS(user_id)
);

CREATE TABLE CONFIG (
	config_id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER,
	config_parameter TEXT NOT NULL UNIQUE,
	config_value TEXT NOT NULL,
	config_lastchanged DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES USERS(user_id)
);

CREATE TABLE INVITES (
	invite_id INTEGER PRIMARY KEY AUTOINCREMENT,
	invite_token TEXT NOT NULL UNIQUE,
	invite_issued DATETIME DEFAULT CURRENT_TIMESTAMP,
	invite_expiry DATETIME,
	invite_claimed_by TEXT,
	invite_claimed BOOLEAN DEFAULT FALSE
);

CREATE TABLE RESETS (
	reset_id INTEGER PRIMARY KEY AUTOINCREMENT,
	reset_email TEXT NOT NULL UNIQUE,
	reset_token TEXT NOT NULL UNIQUE,
	reset_issued TEXT NOT NULL,
	reset_expiry TEXT NOT NULL,
	reset_used BOOLEAN DEFAULT FALSE
);

CREATE TABLE IPS (
	ip_id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER,
	ip_address TEXT NOT NULL UNIQUE,
	ip_first_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
	ip_last_seen DATETIME,
	ip_days_seen INTEGER DEFAULT 1,
	ip_note TEXT,
	FOREIGN KEY (user_id) REFERENCES USERS(user_id)
);

CREATE TABLE ULOGS (
	ulog_id INTEGER PRIMARY KEY AUTOINCREMENT,
	ulog DATETIME DEFAULT CURRENT_TIMESTAMP,
	user_id INTEGER,
	ulog_event TEXT NOT NULL,
	FOREIGN KEY (user_id) REFERENCES USERS(user_id)
);