package main

import (
	"fmt"
	"go-rest-api/api/model"
	"go-rest-api/api/web"
	"net/http"
	"net/url"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"

	"github.com/guregu/sqlx"
	log "github.com/sirupsen/logrus"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// connect to DB
	db := sqlx.MustConnect("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?parseTime=true&columnsWithAlias=true&loc=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		fmt.Sprintf("%s:%s", os.Getenv("DB_HOST"), os.Getenv("DB_PORT")),
		os.Getenv("DB_NAME"),
		"Asia%2FTokyo",
	))

	// create repository
	repo, err := model.NewSqlxRepository(db)
	if err != nil {
		log.Println("[ERROR] NewSqlxRepository:", err.Error())
		return
	}

	appURL, err := url.Parse(os.Getenv("APP_URL"))
	if err != nil {
		log.Println("[ERROR] url.Parse(os.Getenv(\"APP_URL\"):", err.Error())
		return
	}

	// create handler
	// TODO:err handling to show error message
	api, err := web.NewAPI(repo, appURL)
	if err != nil {
		log.Error("[ERROR] NewAPI:", err.Error())
		return
	}

	// Create router
	r := NewRouter(api)
	// run server
	http.ListenAndServe(":"+os.Getenv("PORT"), r)

}
