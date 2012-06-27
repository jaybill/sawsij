// Copyright <year> <name>. All rights reserved.
// Use of this source code is governed by license 
// that can be found in the LICENSE file.

package {{ .name }}

import (
	"bitbucket.org/jaybill/sawsij/framework"
	"crypto/md5"
	"fmt"
	"io"
	"time"
)

type User struct {
	Id           int64
	Username     string
	PasswordHash string
	FullName     string
	Email        string
	CreatedOn    time.Time
	Role         int64
}
{{/* TODO passwords should be hashed via bcrypt and a framework function, not md5 (issue #13) */}}
func (u *User) SetPassword(password string, salt string) {
	h := md5.New()
	io.WriteString(h, salt)
	io.WriteString(h, password)
	u.PasswordHash = fmt.Sprintf("%x", h.Sum(nil))
}

func (u *User) TestPassword(password string, a *framework.AppScope) (valid bool) {
	valid = false
	salt, _ := a.Config.Get("encryption.salt")

	h := md5.New()
	if salt != "" {
		io.WriteString(h, salt)
	}

	io.WriteString(h, password)
	tHash := fmt.Sprintf("%x", h.Sum(nil))

	if u.PasswordHash == tHash {
		valid = true
	}
	return
}

func (u *User) GetRole() int64 {
	return u.Role
}

func (u *User) ClearPasswordHash() {
	u.PasswordHash = ""
}

