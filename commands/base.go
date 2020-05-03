package commands

import (
	"github.com/sirupsen/logrus"
	"github.com/enigmaquip/gonab/config"
	"github.com/enigmaquip/gonab/db"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	// App is the main hook to run
	App        = kingpin.New("gonab", "A usenet indexer")
	debug      = App.Flag("debug", "Enable Debug mode.").Bool()
	debugdb    = App.Flag("debugdb", "Log Database queries (noisy).").Default("false").Bool()
	configfile = App.Flag("config", "Config file to use").Default("config.json").ExistingFile()
)

// SetupCommands sets up commands
func SetupCommands() {
	rcmd := &ReleasesCommand{}
	rcmd.configure(App)

	gcmd := &GroupCommand{}
	gcmd.configure(App)

	scanner := &ScanCommand{}
	scanner.configure(App)

	server := &ServerCommand{}
	server.configure(App)

	App.Command("createdb", "Create Database and Tables.").Action(createdb)

	bcmd := &BinariesCommand{}
	bcmd.configure(App)

	regexcmd := &RegexImporter{}
	App.Command("importregex", "Import regexes from nzedb").Action(regexcmd.run)
}

func commonInit() (*config.Config, *db.Handle) {
	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	cfg := loadConfig(*configfile)

	dbh := db.NewDBHandle(cfg.DB.Name, cfg.DB.Username, cfg.DB.Password, cfg.DB.Verbose)

	return cfg, dbh
}

func loadConfig(cfile string) *config.Config {
	if len(cfile) == 0 {
		logrus.Infof("No --config_file given.  Using default: %s", *configfile)
		cfile = *configfile
	}

	logrus.Infof("Got config file: %s\n", cfile)
	cfg := config.NewConfig()
	err := cfg.ReadConfig(cfile)
	if err != nil {
		logrus.Fatal(err)
	}

	// Override cfg from flags
	cfg.DB.Verbose = *debugdb
	return cfg
}
