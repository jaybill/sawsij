// Copyright <year> <name>. All rights reserved.
// Use of this source code is governed by license 
// that can be found in the LICENSE file.

package {{ .name }}

import (
	"bitbucket.org/jaybill/sawsij/framework"		
	"time"
)

// User represents an application user in the database. Conforms to the framework.User interface.
// Roles should be specified with the constants in {{.name}}/constants.go 
type User struct {
	Id           int64
	Username     string
	PasswordHash string
	FullName     string
	Email        string
	CreatedOn    time.Time
	Role         int64
}

// SetPassword generates and sets a password hash from a password string and a salt string. 
// Currently uses the hashing algorithm supplied by the framework. (Required by framework.User)
func (u *User) SetPassword(password string, salt string) {	
	u.PasswordHash = framework.PasswordHash(password, salt)
}

// Tests if the supplied password, when hashed, matches the password hash for the referenced user. (Required by framework.User)
func (u *User) TestPassword(password string, a *framework.AppScope) (valid bool) {
	valid = false
	salt, _ := a.Config.Get("encryption.salt")

	if u.PasswordHash == framework.PasswordHash(password, salt) {
		valid = true
	}
	return
}

// Returns the User's role. (Required by framework.User)
func (u *User) GetRole() int64 {
	return u.Role
}

// Sets the password hash on a user struct to empty so it can be super-safely stored in the session. (Required by framework.User)
func (u *User) ClearPasswordHash() {
	u.PasswordHash = ""
}

