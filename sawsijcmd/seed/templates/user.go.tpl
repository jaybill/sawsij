// Copyright <year> <name>. All rights reserved.
// Use of this source code is governed by license
// that can be found in the LICENSE file.

package {{ .name }}

import (
	"bitbucket.org/jaybill/sawsij/framework"
	"bitbucket.org/jaybill/sawsij/framework/model"
	"fmt"
	"log"
	"net/http"
	"strings"
	"strconv"
	"time"
)

// User represents an application user in the database. Conforms to the framework.User interface.
// Roles should be specified with the constants in {{ .name }}/constants.go
type User struct {
	Id           int64
	Username     string
	PasswordHash string
	FullName     *string
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
	if framework.CompareHashAndPassword(u.PasswordHash, password, salt) {
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

// Looks at the data in the user struct and determines if it's valid. Returns an array of errors if it isn't.
func (u *User) GetValidationErrors(a *framework.AppScope) (errors []string) {

	if len(strings.TrimSpace(u.Username)) == 0 {
		errors = append(errors, "Username cannot be blank.")
	}

	if len(strings.TrimSpace(u.Email)) == 0 {
		errors = append(errors, "Email cannot be blank.")
	}

	t := &model.Table{Db: a.Db}
	user := &User{}
	q := model.Query{}

	var users []interface{}
	var err error

	if u.Id == -1 {
		q.Where = fmt.Sprintf("%v = %v", model.MakeDbName("Email"), a.Db.GetQueries().P(1))
		users, err = t.FetchAll(user, q, u.Email)
	} else {
		q.Where = fmt.Sprintf("%v = %v and %v <> %v", model.MakeDbName("Email"), a.Db.GetQueries().P(1), model.MakeDbName("Id"), a.Db.GetQueries().P(2))
		users, err = t.FetchAll(user, q, u.Email, u.Id)
	}

	if err != nil {
		errors = append(errors, "Database error.")
		return
	} else {
		if len(users) > 0 {
			errors = append(errors, "Email address is already in use.")
		}
	}

	if u.Id == -1 {
		q.Where = fmt.Sprintf("%v = %v", model.MakeDbName("Username"), a.Db.GetQueries().P(1))
		users, err = t.FetchAll(user, q, u.Username)
	} else {
		q.Where = fmt.Sprintf("%v = %v and %v <> %v", model.MakeDbName("Username"), a.Db.GetQueries().P(1), model.MakeDbName("Id"), a.Db.GetQueries().P(2))
		users, err = t.FetchAll(user, q, u.Username, u.Id)
	}

	if err != nil {
		errors = append(errors, "Database error.")
		return
	} else {
		if len(users) > 0 {
			errors = append(errors, "Username is already in use.")
		}
	}

	return
}

// Handles the user admin list page.
func UserAdminListHandler(r *http.Request, a *framework.AppScope, rs *framework.RequestScope) (h framework.HandlerResponse, err error) {
	h.Init()

	t := &model.Table{Db: a.Db}
	user := &User{}
	q := model.Query{}
	q.Order = model.MakeDbName("Username")
	users, err := t.FetchAll(user, q)
	if err == nil {
		h.View["users"] = users
	} else {
		h.Redirect = "/error"
	}

	return
}

// Handles the user edit/insert page
func UserAdminEditHandler(r *http.Request, a *framework.AppScope, rs *framework.RequestScope) (h framework.HandlerResponse, err error) {
	h.Init()
	t := &model.Table{Db: a.Db}
	user := &User{}

	user.Id = framework.GetIntId(rs.UrlParamMap["id"])
	if user.Id != -1 {
		err = t.Fetch(user)
		if err != nil {
			log.Print(err)
			h.Redirect = "/error"
			return
		} else {
			h.View["user"] = user
		}
	}

	h.View["roles"] = map[string]int{"member": R_MEMBER, "admin": R_ADMIN}

	if r.Method == "POST" {

		fn := r.FormValue("FullName")
		user.Username = r.FormValue("Username")
		user.FullName = &fn
		user.Email = r.FormValue("Email")

		tRole := r.FormValue("Role")
		tt, _ := strconv.Atoi(tRole)
		user.Role = int64(tt)

		errors := user.GetValidationErrors(a)

		// Password validation has to be done in the handler because the model doesn't know about the confirmation field
		// or that the field is optional if you're not changing it.
		password := strings.TrimSpace(r.FormValue("Password"))
		passwordAgain := strings.TrimSpace(r.FormValue("PasswordAgain"))
		if len(password) > 0 {
			if password != passwordAgain {
				errors = append(errors, "Passwords do not match.")
			} else {
				salt, _ := a.Config.Get("encryption.salt")
				user.SetPassword(password, salt)
			}
		}

		if user.Id == -1 && len(password) < 1 {
			errors = append(errors, "Password cannot be blank.")
		}

		if len(errors) == 0 {
			if user.Id == -1 {
				// This is an insert
				user.CreatedOn = time.Now()
				err = t.Insert(user)
				if err != nil {
					log.Print(err)
					h.Redirect = "/error"
					return
				} else {
					h.View["success"] = "User created."
				}
			} else {
				// This is an update
				err = t.Update(user)
				if err != nil {
					log.Print(err)
					h.Redirect = "/error"
					return
				} else {
					h.View["success"] = "User updated."
				}

			}

		} else {
			h.View["errors"] = errors
		}
		// Pass back marshaled struct, even if it isn't valid, to allow correction of mistakes.
		h.View["user"] = user

	}
	if user.Id != -1 {
		h.View["update"] = true
	}

	return
}

// Handles the user delete page
func UserAdminDeleteHandler(r *http.Request, a *framework.AppScope, rs *framework.RequestScope) (h framework.HandlerResponse, err error) {
	h.Init()

	t := &model.Table{Db: a.Db}
	user := &User{}

	user.Id = framework.GetIntId(rs.UrlParamMap["id"])
	if user.Id != -1 {
		err = t.Fetch(user)
		if err != nil {
			log.Print(err)
			h.Redirect = "/error"
			return
		} else {
			h.View["user"] = user
		}
	} else {
		log.Print("Delete user called without user id.")
		h.Redirect = "/error"
		return
	}

	h.View["user"] = user

	if r.Method == "POST" {
		t.Delete(user)
		h.Redirect = "/admin/users"
	}

	return
}
