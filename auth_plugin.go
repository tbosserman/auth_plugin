package main

import "fmt"
import "flag"
import "log"
import "log/syslog"
import "os"
import "os/user"
import "io/ioutil"
import "net"
import "net/http"
import "encoding/json"
import "path"
import "strings"
import "strconv"

import (
    //#include <unistd.h>
    //#include <errno.h>
    "C"
)

const (
    ConfigFile string = "/usr/local/etc/auth_plugin.json"
    DefaultUnixSocket string = "/var/run/docker/plugins/auth_plugin.sock"
    DefaultLogFile string = "syslog:daemon:info"
)

type Body struct {
    User     		string
    RequestUri		string
    UserAuthNMethod	string
    RequestMethod	string
}

type CliFlags struct {
    cfgfile	*string
    socket	*string
    logfile	*string
}

type CfgOptions struct {
    socket	string
    logfile	string
}

var logger	*log.Logger
var options	CfgOptions
var progname	string

var levels = map[string]syslog.Priority{
    "debug":	syslog.LOG_DEBUG,
    "info":	syslog.LOG_INFO,
    "notice":	syslog.LOG_NOTICE,
    "warning":	syslog.LOG_WARNING,
    "err":	syslog.LOG_ERR,
    "crit":	syslog.LOG_CRIT,
    "alert":	syslog.LOG_ALERT,
    "emerg":	syslog.LOG_EMERG,
}

var facilities = map[string]syslog.Priority{
    "kern":	syslog.LOG_KERN,
    "user":	syslog.LOG_USER,
    "mail":	syslog.LOG_MAIL,
    "daemon":	syslog.LOG_DAEMON,
    "auth":	syslog.LOG_AUTH,
    "syslog":	syslog.LOG_SYSLOG,
    "lpr":	syslog.LOG_LPR,
    "news":	syslog.LOG_NEWS,
    "uucp":	syslog.LOG_UUCP,
    "cron":	syslog.LOG_CRON,
    "authpriv":	syslog.LOG_AUTHPRIV,
    "ftp":	syslog.LOG_FTP,
    "local0":	syslog.LOG_LOCAL0,
    "local1":	syslog.LOG_LOCAL1,
    "local2":	syslog.LOG_LOCAL2,
    "local3":	syslog.LOG_LOCAL3,
    "local4":	syslog.LOG_LOCAL4,
    "local5":	syslog.LOG_LOCAL5,
    "local6":	syslog.LOG_LOCAL6,
    "local7":	syslog.LOG_LOCAL7,
}

func usage() {
    fmt.Fprintf(os.Stderr,
      "usage: auth_plugin [-c config_file] [-l logfile] [-s socket]\n")
    fmt.Fprintf(os.Stderr, "       default config file is %s\n", ConfigFile)
    fmt.Fprintf(os.Stderr, "       default socket is %s\n", DefaultUnixSocket)
    fmt.Fprintf(os.Stderr, "       default logfile is %s\n", DefaultLogFile)
    os.Exit(1)
}

func myHandler(writer http.ResponseWriter, reader *http.Request) {
    var data		[]byte
    var err		error
    var jdata		Body
    var user, uri	string
    var method		string

    // logger.Printf("Method: %s\n", reader.Method)
    // logger.Printf("Path: %s\n", reader.URL.Path)
    // logger.Printf("User agent: %s\n", reader.UserAgent())
    // reader.Body is an io.ReadCloser
    if data, err = ioutil.ReadAll(reader.Body); err != nil {
	logger.Fatal(err)
    }
    //logger.Printf("Body: %s\n", data)

    user = "(docker.sock)"
    uri = reader.URL.Path
    method = "(unknown)"
    if len(data) > 0 {
	if  err = json.Unmarshal(data, &jdata); err != nil {
	    logger.Printf("Error parsing JSON body: %s\n", err)
	    return
	}
	if jdata.User != "" {
	    user = jdata.User
	}
	uri = jdata.RequestUri
	method = jdata.RequestMethod
    }
    if reader.URL.Path == "/AuthZPlugin.AuthZReq" && uri != "/_ping" {
	logger.Printf("%s %s %s\n", user, method, uri)
    }

    //logger.Print("========== END OF REQUEST ==========\n")

    if reader.Method != "POST" {
	return
    }

    if reader.URL.Path == "/Plugin.Activate" {
	data = []byte("{ \"Implements\": [\"authz\"] }")
	writer.Write(data)
    } else if reader.URL.Path == "/AuthZPlugin.AuthZReq" ||
              reader.URL.Path == "/AuthZPlugin.AuthZRes" {
	data = []byte("{  \"Allow\": true }")
	writer.Write(data)
    }

}

func parse_options() {
    var flags		CliFlags
    var data		[]byte
    var err		error

    // Process command line arguments. These will override config file.
    flag.Usage = usage
    flags.socket = flag.String("s", "", "Unix Socket")
    flags.logfile = flag.String("l", "", "Log file")
    flags.cfgfile = flag.String("c", ConfigFile, "Config file")
    flag.Parse()

    // Read the config file if it exists.
    if _, err = os.Stat(*flags.cfgfile); os.IsExist(err) {
	if data, err = ioutil.ReadFile(*flags.cfgfile); err != nil {
	    fmt.Fprintf(os.Stderr, "error opening %s for reading: %s\n",
	        flags.cfgfile, err)
	    os.Exit(1)
	}
    }
    if len(data) > 0 {
	if  err = json.Unmarshal(data, &options); err != nil {
	    fmt.Fprintf(os.Stderr, "Error parsing JSON body: %s\n", err)
	    os.Exit(1)
	}
    }

    // Let the CLI options override the config file.
    // Use defaults where nothing was specified in either place.
    if *flags.socket != "" {
	options.socket = *flags.socket
    }
    if options.socket == "" {
	options.socket = DefaultUnixSocket
    }

    if *flags.logfile != "" {
	options.logfile = *flags.logfile
    }
    if options.logfile == "" {
	options.logfile = DefaultLogFile
    }
}

func openlog() {
    var fd		*os.File
    var flags		int
    var loginfo		[]string
    var ok		bool
    var fac, lev, pri	syslog.Priority
    var writer		*syslog.Writer
    var err		error

    loginfo = strings.Split(options.logfile, ":")
    if len(loginfo) == 3 && loginfo[0] == "syslog" {
	if fac, ok = facilities[loginfo[1]]; !ok {
	    fmt.Fprintf(os.Stderr, "invalid facility: '%s'\n", loginfo[1])
	    os.Exit(1)
	}
	if lev, ok = levels[loginfo[2]]; !ok {
	    fmt.Fprintf(os.Stderr, "invalid level: '%s'\n", loginfo[2])
	    os.Exit(1)
	}
	pri = fac | lev
	if writer, err = syslog.New(pri, progname); err != nil {
	    fmt.Fprintln(os.Stderr, err)
	    os.Exit(1)
	}
	logger = log.New(writer, "", 0)
    } else {
	flags = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	fd, err = os.OpenFile(options.logfile, flags, 0600); if err != nil {
	    fmt.Fprintf(os.Stderr, "error opening %s for writing: %s\n",
		options.logfile, err)
	    os.Exit(1)
	}
	logger = log.New(fd, "", log.LstdFlags)
    }
    logger.Print("auth_plugin starting up...\n")
}

func main() {
    var err		error
    var listener	net.Listener
    var runas		*user.User
    var uid		int

    progname = path.Base(os.Args[0])
    parse_options()
    openlog()

    os.Remove(options.socket)
    listener, err = net.Listen("unix", options.socket)
    if err != nil {
	logger.Fatal(err)
    }

    // We started as root, now drop privileges to run as "nobody"
    if runas, err = user.Lookup("nobody"); err != nil {
	logger.Println(err)
	os.Exit(1)
    }
    uid, _ = strconv.Atoi(runas.Uid)
    if cerr, errno := C.setuid(C.__uid_t(uid)); cerr != 0 {
	logger.Printf("setuid returned errno %d", errno)
    }

    http.HandleFunc("/", myHandler)
    logger.Fatal(http.Serve(listener, nil))
}
