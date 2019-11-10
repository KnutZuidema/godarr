package main

import (
	"flag"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"github.com/rubenv/sql-migrate"

	"github.com/KnutZuidema/godarr/pkg/api"
	"github.com/KnutZuidema/godarr/pkg/database"
	"github.com/KnutZuidema/godarr/pkg/model"
)

func main() {
	var (
		serverAddress   = flag.String("server.address", "localhost:5000", "address the server should listen on")
		postgresAddress = flag.String("postgres.address", "postgres://postgres@localhost/postgres?sslmode=disable", "address for the postgres database")
		postgresMigrate = flag.Bool("postgres.migrate", true, "whether to execute migrations, default true")
	)
	flag.Parse()
	sqlxDB, err := sqlx.Open("postgres", *postgresAddress)
	if err != nil {
		logrus.Fatal("database open: ", err)
	}
	if *postgresMigrate {
		migrations := migrate.FileMigrationSource{
			Dir: "migrations",
		}
		n, err := migrate.Exec(sqlxDB.DB, "postgres", migrations, migrate.Up)
		if err != nil {
			logrus.Fatal("migrate: ", err)
		}
		logrus.Infof("executed %d migrations", n)
	}
	db, err := database.New(sqlxDB)
	if err != nil {
		logrus.Fatal("initialize database: ", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logrus.Error("database close: ", err)
		}
	}()
	addedItems := make(chan model.Item)
	server := api.NewServer(db, addedItems, nil)
	logrus.Infof("Listening on %s", *serverAddress)
	if err := http.ListenAndServe(*serverAddress, server.Router); err != nil {
		logrus.Fatal("")
	}
}
