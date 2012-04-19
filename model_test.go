package sawsij

import(    
    "testing"
    "time"
    "flag"
    
    "github.com/stathat/jconfig"
    
)

var config *jconfig.Config

type Post struct{
    Id          int64
    Title       string
    Body        string
    CreatedOn   time.Time
}



func configure(){
    var configFile string
    
    flag.StringVar(&configFile,"c", "./config.json", "path to config file")

    flag.Parse()

    config = jconfig.LoadConfig(configFile)
    
}

func TestInsert(t *testing.T){


    configure()
    
    
}
