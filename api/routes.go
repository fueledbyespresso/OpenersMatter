package api

import (
	"OpenersMatter/database"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/mmcloughlin/geohash"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// Routes All the routes created by the package nested in
// api/v1/*
func Routes(r *gin.RouterGroup, db *database.DB) {
	resourceRoute(r, db)
}

func resourceRoute(r *gin.RouterGroup, db *database.DB) {
	r.GET("/concertsO", overlappedContent(db))
	r.GET("/concerts", getConcerts(db))
}

func overlappedContent(db *database.DB) gin.HandlerFunc {
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
		getTopSpotifyArtists(&artists, accessToken)

		c.JSON(200, "test")
	}
}

type following struct {
	Artists resp `json:"artists"`
}

type resp struct {
	Href  string   `json:"href"`
	Items []artist `json:"items"`
}
type artist struct {
	Name string `json:"name"`
}

func getTopSpotifyArtists(artists *resp, accessToken string) {
	req, _ := http.NewRequest("GET", "https://api.spotify.com/v1/me/top/artists?limit=50", nil)
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
}

func getFollowedArtists(followingArtists *following, accessToken string) {
	req, _ := http.NewRequest("GET", "https://api.spotify.com/v1/me/following?type=artist&limit=50", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	client := &http.Client{}
	response, err := client.Do(req)
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(contents, &followingArtists)
	if err != nil {
		log.Println(err)
	}
}

func getConcerts(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var artists resp
		var followingArtist following
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
		getTopSpotifyArtists(&artists, accessToken)
		getFollowedArtists(&followingArtist, accessToken)
		favArtists := make(map[string]string)
		for _, item := range artists.Items {
			favArtists[strings.ToLower(item.Name)] = ""
		}
		for _, item := range followingArtist.Artists.Items {
			favArtists[strings.ToLower(item.Name)] = ""
		}
		events := getTicketmasterConcerts()
		for _, compactEvent := range events {
			for _, artistPlaying := range compactEvent.Attractions {
				if _, ok := favArtists[strings.ToLower(artistPlaying)]; ok {
					favArtists[strings.ToLower(artistPlaying)] = compactEvent.Name
				}
			}
		}
		c.JSON(200, favArtists)
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

func getTicketmasterConcerts() []CompactEvent {
	var shows Events
	var shows2 Events

	seattle := geohash.Encode(47.5, -122)
	seattle = seattle[:5]
	resp, err := http.Get("https://app.ticketmaster.com/discovery/v2/events.json?size=200&segmentName=Music&geoPoint=" + seattle + "&apikey=" + os.Getenv("TICKETMASTER_KEY"))
	if err != nil {
		return []CompactEvent{}
	}
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []CompactEvent{}
	}
	err = json.Unmarshal(contents, &shows)
	if err != nil {
		log.Println(err)
	}

	resp2, err := http.Get("https://app.ticketmaster.com/discovery/v2/events.json?size=200&segmentName=Music&geoPoint=" + seattle + "&apikey=" + os.Getenv("TICKETMASTER_KEY"))
	if err != nil {
		return []CompactEvent{}
	}
	contents2, err := ioutil.ReadAll(resp2.Body)
	if err != nil {
		return []CompactEvent{}
	}
	err = json.Unmarshal(contents2, &shows2)
	if err != nil {
		log.Println(err)
	}
	redEvents := reduceEvents(shows)
	redEvents2 := append(redEvents, reduceEvents(shows2)...)

	return redEvents2
}
