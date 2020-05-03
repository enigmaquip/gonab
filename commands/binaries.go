package commands

import (
	"github.com/enigmaquip/gonab/db"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

type BinariesCommand struct{}

func (s *BinariesCommand) configure(app *kingpin.Application) {
	app.Command("makebinaries", "Create binaries from parts").Action(s.run)
}

func (s *BinariesCommand) run(c *kingpin.ParseContext) error {
	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	cfg := loadConfig(*configfile)

	dbh := db.NewDBHandle(cfg.DB.Name, cfg.DB.Username, cfg.DB.Password, cfg.DB.Host, cfg.DB.Verbose)
	return dbh.MakeBinaries()
}
