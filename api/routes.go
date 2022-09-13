package api

import (
	"OpenersMatter/database"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// Routes All the routes created by the package nested in
// api/v1/*
func Routes(r *gin.RouterGroup, db *database.DB) {
	resourceRoute(r, db)
}

func resourceRoute(r *gin.RouterGroup, db *database.DB) {
	r.GET("/concerts", getConcerts(db))
	r.GET("/getconcerts", getTicketmasterConcerts())
}

type resp struct {
	Href  string   `json:"href"`
	Items []artist `json:"items"`
}
type artist struct {
	Name string `json:"name"`
}

func getConcerts(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var artists resp
		session, err := db.SessionStore.Get(c.Request, "session")
		if err != nil {
			c.AbortWithStatusJSON(500, "The server was unable to retrieve this session")
			return
		}
		spotifyID := session.Values["SpotifyID"]

		accessTokenRow, err := db.Db.Query(`SELECT access_token from account WHERE spotify_id=$1`, spotifyID)
		if err != nil {
			database.CheckDBErr(err.(*pq.Error), c)
			return
		}
		var accessToken string
		for accessTokenRow.Next() {
			err = accessTokenRow.Scan(&accessToken)
			if err != nil {
				c.AbortWithStatusJSON(500, "The server was unable to retrieve school info")
			}
		}

		req, _ := http.NewRequest("GET", "https://api.spotify.com/v1/me/top/artists", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		client := &http.Client{}
		response, err := client.Do(req)
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return
		}
		err = json.Unmarshal(contents, &artists)
		if err != nil {
			log.Println(err)
		}
		c.JSON(200, artists)
	}
}

type Events struct {
	Embedded struct {
		Events []event `json:"events"`
	} `json:"_embedded"`
}

type event struct {
	Name        string `json:"name"`
	Attractions struct {
		Attractions []attraction
	} `json:"_embedded"`
}

type attraction struct {
	Name string `json:"name"`
}

type CompactEvent struct {
	Name        string `json:"name"`
	Attractions []string
}

func reduceEvents(events Events) []CompactEvent {
	var compactEvents []CompactEvent

	for _, event := range events.Embedded.Events {
		var attractions []string
		for _, attraction := range event.Attractions.Attractions {
			attractions = append(attractions, attraction.Name)
		}
		formatted := CompactEvent{
			Name:        event.Name,
			Attractions: attractions,
		}

		compactEvents = append(compactEvents, formatted)
	}

	return compactEvents
}

func getTicketmasterConcerts() gin.HandlerFunc {
	return func(c *gin.Context) {
		var shows Events
		resp, err := http.Get("https://app.ticketmaster.com/discovery/v2/events.json?size=100&city=seattle&radius=100&units=mi&apikey=" + os.Getenv("TICKETMASTER_KEY"))
		if err != nil {
			return
		}
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}
		err = json.Unmarshal(contents, &shows)
		if err != nil {
			log.Println(err)
		}
		c.JSON(200, reduceEvents(shows))
	}
}
