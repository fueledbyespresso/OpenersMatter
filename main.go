package main

import (
	"OpenersMatter/api"
	authentication "OpenersMatter/auth"
	"OpenersMatter/database"
	"fmt"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"net/http"
	"os"
	"time"
)

//Load the environment variables from the projectvars.env file
func initEnv() {
	if _, err := os.Stat("projectvars.env"); err == nil {
		err = godotenv.Load("projectvars.env")
		if err != nil {
			fmt.Println("Error loading environment.env")
		}
		fmt.Println("Current environment:", os.Getenv("ENV"))
	}
}

func forceSSL() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header.Get("x-forwarded-proto") != "https" {
			sslUrl := "https://" + c.Request.Host + c.Request.RequestURI
			c.Redirect(http.StatusTemporaryRedirect, sslUrl)
			return
		}
		c.Next()
	}
}

func createServer(dbConnection *database.DB) *gin.Engine {
	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	if os.Getenv("ENV") != "DEV" {
		r.Use(forceSSL())
	}

	authentication.Routes(r.Group("oauth/v1"), dbConnection)
	api.Routes(r.Group("api/v1"), dbConnection)
	r.Use(static.Serve("/", static.LocalFile("./frontend/build", true)))

	return r
}

func main() {
	initEnv()
	database.PerformMigrations("file://database/migrations")
	authentication.ConfigOauth()
	db := database.InitDBConnection()
	defer db.Close()

	SStore := database.InitOauthStore()
	// Run a background goroutine to clean up expired sessions from the database.
	defer SStore.StopCleanup(SStore.Cleanup(time.Minute * 5))
	dbConnection := &database.DB{Db: db, SessionStore: SStore}

	r := createServer(dbConnection)

	_ = r.Run()
}
