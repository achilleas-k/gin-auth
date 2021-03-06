// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package data

import (
	"database/sql"
	"time"

	"github.com/G-Node/gin-auth/conf"
	"github.com/G-Node/gin-auth/util"
)

// Session contains data about session tokens used to identify
// logged in accounts.
type Session struct {
	Token       string
	Expires     time.Time
	AccountUUID string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ListSessions returns all sessions sorted by creation time.
func ListSessions() []Session {
	const q = `SELECT * FROM Sessions WHERE expires > now() ORDER BY createdAt`

	sessions := make([]Session, 0)
	err := database.Select(&sessions, q)
	if err != nil {
		panic(err)
	}

	return sessions
}

// GetSession returns a session with a given token.
// Returns false if no such session exists.
func GetSession(token string) (*Session, bool) {
	const q = `SELECT * FROM Sessions WHERE token=$1 AND expires > now()`

	session := &Session{}
	err := database.Get(session, q, token)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return session, err == nil
}

// Create stores a new session.
// If the token is empty a random token will be generated.
func (sess *Session) Create() error {
	const q = `INSERT INTO Sessions (token, expires, accountUUID, createdAt, updatedAt)
	           VALUES ($1, $2, $3, now(), now())
	           RETURNING *`

	sess.Expires = time.Now().Add(conf.GetServerConfig().SessionLifeTime)
	if sess.Token == "" {
		sess.Token = util.RandomToken()
	}

	return database.Get(sess, q, sess.Token, sess.Expires, sess.AccountUUID)
}

// UpdateExpirationTime updates the expiration time and stores
// the new time in the database.
func (sess *Session) UpdateExpirationTime() error {
	const q = `UPDATE Sessions SET (expires, updatedAt) = ($1, now())
	           WHERE token=$2
	           RETURNING *`

	return database.Get(sess, q, time.Now().Add(conf.GetServerConfig().SessionLifeTime), sess.Token)
}

// Delete removes a session from the database.
func (sess *Session) Delete() error {
	const q = `DELETE FROM Sessions WHERE token=$1`

	_, err := database.Exec(q, sess.Token)
	return err
}
