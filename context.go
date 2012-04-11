package sawsij

import(	
	"launchpad.net/mgo"	
	"github.com/stathat/jconfig"
)

type Context struct{
	Config 		*jconfig.Config	
	DbSession	*mgo.Session	
}
