package sawsij

import (
	"database/sql"
	"github.com/stathat/jconfig"
)

type Context struct {
	Config   *jconfig.Config
	Db       *sql.DB
	BasePath string
}

