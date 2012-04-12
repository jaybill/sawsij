package sawsij

import (
	"github.com/stathat/jconfig"
	"launchpad.net/mgo"
)

type Context struct {
	Config      *jconfig.Config
	DbSession   *mgo.Session
	BasePath    string

}
