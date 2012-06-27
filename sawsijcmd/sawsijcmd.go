package main

import (
	"bitbucket.org/jaybill/sawsij/framework"
	"database/sql"
	"fmt"
	"io"
	"crypto/md5"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

var env map[string]string
var goroot string = "/usr/local/go"
var gopath string = ""
var confdir string = ""
var envfile string = ""
var gobinpath string = ""
var sawsijhome string = ""

func main() {
	var err error

	envStrings := os.Environ()

	env = make(map[string]string)

	for _, envString := range envStrings {
		keyval := strings.Split(envString, "=")
		env[keyval[0]] = keyval[1]
	}

	confdir = env["HOME"] + "/.sawsij"
	envfile = confdir + "/sawsijenv"

	if env["SAWSIJ_HOME"] == "" {
		fmt.Println("SAWSIJ_HOME environment variable is not set.")
		os.Exit(1)
	} else {
		sawsijhome = env["SAWSIJ_HOME"]
	}

	if env["GOPATH"] == "" {
		fmt.Println("GOPATH environment variable is not set.")
		os.Exit(1)
	} else {
		gopath = env["GOPATH"]
	}

	gobinpath, err = exec.LookPath("go")
	if err != nil {
		fmt.Println("Can't find go. Is it installed and in your path?")
	} else {

		version, err := exec.Command(gobinpath, "version").Output()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			fmt.Printf("Using %s\n", gobinpath)
			fmt.Printf("%s\n", strings.TrimSpace(string(version)))
		}

	}

	if len(os.Args) == 1 {
		fmt.Println("No command specified.")
		os.Exit(1)
	}
	command := strings.TrimSpace(os.Args[1])

	if env["SAWSIJ_SETUP"] != "1" {
		envSetup()
	}

	switch command {
	case "new":
		new()	
	case "crudify":
		// TODO create DAL and CRUD based on database table (issue #12)
	default:
		fmt.Printf("Command %q not recognized.\n", command)
		os.Exit(1)
	}

}

func envSetup() {

	envt := `# This file contains sawsij environment variables for development
# Be sure to add the line "source $HOME/.sawsij/sawsijenv" to the end of ` + env["HOME"] + `/.profile
# so this will be read.

SAWSIJ_SETUP=1
export SAWSIJ_SETUP
`

	_, err := os.Open(confdir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(confdir, os.FileMode(0744))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			} else {
				fmt.Printf("Created new sawsij config dir %q\n", confdir)
			}
		}
	} else {
		fmt.Printf("Found config dir %q\n", confdir)
	}

	_, err = os.Open(envfile)
	if err != nil {
		if os.IsNotExist(err) {
			err = framework.WriteStringToFile(envt, envfile)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			} else {
				fmt.Println("****************************\n** NOTICE! ACTION NEEDED! **\n****************************")
				fmt.Println("A new file containing sawsij environment variables was created.")
				fmt.Printf("You must add the line \"source $HOME/.sawsij/sawsijenv\" to the end of %v\n\n", env["HOME"]+"/.profile")
			}
		}
	} else {
		fmt.Printf("Found env file %q\n", envfile)
	}
}

func new() {
	var err error
	var itWorked bool = true

	config := make(map[string]string)
	name := ""
	fmt.Println("****************************\n** CREATE NEW APPLICATION **\n****************************")
	for name == "" {
		name, _ = framework.GetUserInput("Application name", name)
	}

	config["name"] = name
	appserver := name + "server"
	path, _ := os.Getwd()
	path = strings.TrimSpace(path) + "/" + name
	path, _ = framework.GetUserInput("Application path", path)

	config["port"], _ = framework.GetUserInput("Server Port", "8078")
	config["admin_email"], _ = framework.GetUserInput("Admin Email", name + "@" + name + ".com")
	password := ""
	for password == "" {
		password, _ = framework.GetUserInput("Admin Password", password)
	}
	

	fmt.Println("****************************\n** DATABASE CONFIGURATION **\n****************************")

	config["driver"], _ = framework.GetUserInput("Database driver", "postgres")
	dbname, _ := framework.GetUserInput("Database name", name)
	config["schema"], _ = framework.GetUserInput("Database schema", name)
	dbuser, _ := framework.GetUserInput("Database user", name)
	dbpass, _ := framework.GetUserInput("Database password", "")
	dbssl, _ := framework.GetUserInput("Database SSL Mode", "disable")

	config["connect"] = fmt.Sprintf("user=%v password=%v dbname=%v sslmode=%v", dbuser, dbpass, dbname, dbssl)

	config["salt"] = framework.MakeRandomId()
	config["key"] = framework.MakeRandomId()
	
	
	// TODO passwords should be hashed via bcrypt and a framework function, not md5 (issue #13)
	h := md5.New()
	io.WriteString(h, config["salt"])
	io.WriteString(h, password)
	config["password_hash"]  = fmt.Sprintf("%x", h.Sum(nil))

	fmt.Printf("Creating new sawsij app %q in location %v\n", name, path)

	seeddir := sawsijhome + "/seed"

	fmt.Printf("seed dir: %v\n", seeddir)

	tplDir := path + "/templates"
	binDir := path + "/bin"
	srcDir := path + "/src"
	etcDir := path + "/etc"
	pkgDir := path + "/pkg"
	sqlChgDir := path + "/sql/changes"
	sqlObjDir := path + "/sql/objects"

	appdirs := []string{
		tplDir,
		srcDir,
		srcDir + "/" + appserver,
		srcDir + "/" + name,		
		etcDir,
		pkgDir,
		sqlChgDir,
		sqlObjDir}

	
	_, err = os.Stat(path)

	if !os.IsNotExist(err) {
		fmt.Printf("Target directory %v already exists.\n", path)
		os.Exit(1)
	}

	for _, adir := range appdirs {
		err = os.MkdirAll(adir, os.FileMode(0744))
		if err != nil {
			fmt.Println(err)

		} else {
			err = os.MkdirAll(binDir, os.FileMode(0755))
			if err != nil {
				fmt.Println(err)
				itWorked = false
			}

		}
	}

	var tpls = []struct {
		Name string
		Dest string
	}{
		{"admin.html.tpl", path + "/templates/admin.html"},
		{"admin-footer.html.tpl", path + "/templates/admin-footer.html"},
		{"admin-header.html.tpl", path + "/templates/admin-header.html"},
		{"appserver.go.tpl", path + "/src/" + appserver + "/" + appserver + ".go"},
		{"config.yaml.tpl", path + "/etc/config.yaml"},
		{"dbversions.yaml.tpl", path + "/etc/dbversions.yaml"},
		{"constants.go.tpl", path + "/src/" + name + "/constants.go"},
		{"footer.html.tpl", path + "/templates/footer.html"},
		{"header.html.tpl", path + "/templates/header.html"},
		{"index.html.tpl", path + "/templates/index.html"},
		{"license.tpl", path + "/LICENSE"},
		{config["driver"] + "_0001.sql.tpl", path + "/sql/changes/" + config["driver"] +  "_" + config["schema"] + "_0001.sql"},
		{config["driver"] + "_views.sql.tpl", path + "/sql/objects/" + config["driver"] + "_" + config["schema"] + "_views.sql"},
		{"user.go.tpl", path + "/src/" + name + "/user.go"},
	}
	if itWorked {
		for _, t := range tpls {
			fmt.Printf("Parsing template %v\n", t.Name)
			pt, err := template.New(t.Name).ParseFiles(seeddir + "/templates/" + t.Name)

			if err != nil {
				fmt.Println(err)
				itWorked = false
			}

			f, err := os.Create(t.Dest)
			if err != nil {
				fmt.Println(err)
				itWorked = false
			} else {
				defer f.Close()

				err = pt.ExecuteTemplate(f, t.Name, config)
				if err != nil {
					fmt.Println(err)
					itWorked = false
				}
			}

		}

		err = framework.CopyDir(seeddir+"/static", path+"/static")
		if err != nil {
			fmt.Println(err)
			itWorked = false
		}
	}

	
	if itWorked {
		db, err := sql.Open(config["driver"], config["connect"])
		if err != nil {
			fmt.Println(err)
			itWorked = false
		}
		
		// TODO Remove hardcoded sql string, replace with driver based lookup (issue #11)
		tcq := "SELECT count(*) as tables FROM information_schema.tables WHERE table_schema = $1;"
		row := db.QueryRow(tcq, config["schema"])
		tcount := 0

		err = row.Scan(&tcount)
		if err != nil {
			fmt.Println(err)
			itWorked = false
		} else {
			if tcount > 0 {
				fmt.Println("Specified database/schema is not empty.")
				itWorked = false
			}
		}

		if itWorked {
			dbscript := path + "/sql/changes/" + config["driver"] +  "_" + config["schema"] + "_0001.sql"
			fmt.Printf("Running db script: %v\n", dbscript)
			bQuery, err := ioutil.ReadFile(dbscript)
			if err != nil {
				fmt.Println(err)
			} else {
				sQuery := string(bQuery)
				queries := strings.Split(sQuery, ";")
				
				for _, query := range queries {
					query = strings.TrimSpace(query)
					if query != "" {						
						_, err = db.Exec(query)
						if err != nil {
							fmt.Println(err)
							itWorked = false
						}
					}
				}
			}
		}
	}

	if itWorked {

		ef := `
# start %v environment
if [ "$%v" == '' ]; then
	export %v=%v
	GOPATH=$GOPATH:$%v
	PATH=$PATH:$%v/bin
	export GOPATH
	export PATH
fi
# end %v environment
`
		ucname := strings.ToUpper(name)
		envname := "SJ_" + ucname
		ef = fmt.Sprintf(ef, name, envname, envname, path, envname, envname, name)

		err = framework.AppendStringToFile(ef, envfile)
		if err != nil {
			fmt.Println(err)
			itWorked = false
		}

		// set environment variables so we can build the application. 
		err = os.Setenv(envname, path)
		if err != nil {
			fmt.Println(err)
			itWorked = false
		}
		err = os.Setenv("GOPATH", env["GOPATH"]+":"+path)
		if err != nil {
			fmt.Println(err)
			itWorked = false
		}
		err = os.Setenv("PATH", env["PATH"]+":"+path+"/bin")
		if err != nil {
			fmt.Println(err)
			itWorked = false
		}

	}

	// "go install" new app

	cm := `
*************************
**  APPLICATION READY  **
*************************

Your application has been built and lives in %v

New environment variables have been added, so be 
sure to issue to following command before continuing:

source ~/.sawsij/sawsijenv

To start your app:

cd %v
%v .		

(Use CRTL + C to stop it.)

You can then point a browser at http://localhost:%v

The admin panel is at http://localhost:%v/admin
Your username is "admin" and your password is the one you chose above.
`
	if itWorked {
		fmt.Printf("%v %v %v\n", gobinpath, "install", appserver)
		compileMessage, err := exec.Command(gobinpath, "install", appserver).Output()

		if err != nil {
			fmt.Println(compileMessage)
			fmt.Println(err)
			itWorked = false
		} else {

			fmt.Printf(cm, path, path, appserver, config["port"],config["port"])
		}
	}

	if !itWorked {
		err = os.RemoveAll(path)
		if err != nil {
			fmt.Println(err)
		}
		failMsg := `
************************
**    BUILD FAILED    **
************************

Your application could not be built. Please see above for detailed error messages.

`
		fmt.Println(failMsg)
		os.Exit(1)

	}

}
