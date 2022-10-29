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
	"strconv"
	"strings"
)

// Routes All the routes created by the package nested in
// api/v1/*
func Routes(r *gin.RouterGroup, db *database.DB) {
	resourceRoute(r, db)
}

func resourceRoute(r *gin.RouterGroup, db *database.DB) {
	r.GET("/concerts", getConcerts(db))
}

type followingJSONResponse struct {
	Artists artistsJSONField `json:"artists"`
}

type artistsJSONField struct {
	Href  string `json:"href"`
	Items []struct {
		Name   string `json:"name"`
		Images []struct {
			Height int
			Width  int
			URL    string
		} `json:"images"`
	} `json:"items"`
}

func getTopSpotifyArtists(artists *artistsJSONField, accessToken string) {
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

func getFollowedArtists(followingArtists *followingJSONResponse, accessToken string) {
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

func getImportantEvents(accessToken string, longStr string, latStr string) []events {
	var topArtists artistsJSONField
	var followingArtist followingJSONResponse

	getTopSpotifyArtists(&topArtists, accessToken)
	getFollowedArtists(&followingArtist, accessToken)
	allArtists := make(map[string]bool)
	for _, item := range topArtists.Items {
		allArtists[strings.ToLower(item.Name)] = true
	}
	for _, item := range followingArtist.Artists.Items {
		allArtists[strings.ToLower(item.Name)] = true
	}
	allEvents := getTicketmasterConcerts(longStr, latStr)
	var curatedEvents []events
	for _, event := range allEvents {
		hasFavoriteArtist := false
		for playingArtist := range event.Attractions {
			if _, isPlaying := allArtists[strings.ToLower(playingArtist)]; isPlaying {
				event.Attractions[playingArtist] = true
				hasFavoriteArtist = true
			}
		}
		if hasFavoriteArtist {
			curatedEvents = append(curatedEvents, event)
		}
	}
	return curatedEvents
}

func getConcerts(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
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
		longitude := c.DefaultQuery("long", "0")
		latitude := c.DefaultQuery("lat", "0")

		allEvents := getImportantEvents(accessToken, longitude, latitude)
		c.JSON(200, allEvents)
	}
}

type eventsResponseJSON struct {
	Embedded struct {
		Events []eventFieldJSON `json:"events"`
	} `json:"_embedded"`
}

type eventFieldJSON struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Images []struct {
		Ratio  string `json:"ratio"`
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"images"`
	AgeRestrictions struct {
		LegalAgeEnforced bool `json:"legalAgeEnforced"`
	} `json:"ageRestrictions"`
	Attractions struct {
		Attractions []attractionFieldJSON `json:"attractions"`
		Venues      []venuesFieldJSON     `json:"venues"`
	} `json:"_embedded"`
	Dates struct {
		Start struct {
			LocalDate string `json:"localDate"`
		} `json:"start"`
	} `json:"dates"`
}

type attractionFieldJSON struct {
	Name string `json:"name"`
}

type venuesFieldJSON struct {
	Name string `json:"name"`
	City struct {
		Name string `json:"name"`
	} `json:"city"`
}

type events struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Images []struct {
		Ratio  string `json:"ratio"`
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"images"`
	AgeRestricted bool            `json:"ageRestricted"`
	Attractions   map[string]bool `json:"attractions"`
	StartDate     string          `json:"startDate"`
}

func removeExcessDetails(eventsJSON eventsResponseJSON) []events {
	var compactEvents []events

	for _, event := range eventsJSON.Embedded.Events {
		var attractions []string
		for _, attraction := range event.Attractions.Attractions {
			attractions = append(attractions, attraction.Name)
		}
		formatted := events{
			Name:          event.Name,
			URL:           event.URL,
			Attractions:   make(map[string]bool),
			AgeRestricted: event.AgeRestrictions.LegalAgeEnforced,
			StartDate:     event.Dates.Start.LocalDate,
		}
		for _, attraction := range attractions {
			formatted.Attractions[attraction] = false
		}
		for _, image := range event.Images {
			formatted.Images = append(formatted.Images, image)
		}

		compactEvents = append(compactEvents, formatted)
	}
	return compactEvents
}

func getTicketmasterConcerts(longStr string, latStr string) []events {
	var allEvents []events
	longitude := float64(0)
	latitude := float64(0)

	longitude, _ = strconv.ParseFloat(longStr, 32)
	latitude, _ = strconv.ParseFloat(latStr, 32)
	seattle := geohash.Encode(latitude, longitude)
	seattle = seattle[:5]

	for i := 0; i < 3; i++ {
		var eventsJSON eventsResponseJSON

		resp, err := http.Get("https://app.ticketmaster.com/discovery/v2/events.json?size=200&page=" + strconv.Itoa(i) + "&segmentName=Music&geoPoint=" + seattle + "&apikey=" + os.Getenv("TICKETMASTER_KEY"))
		if err != nil {
			return []events{}
		}
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return []events{}
		}
		err = json.Unmarshal(contents, &eventsJSON)
		if err != nil {
			log.Println(err)
		}

		allEvents = append(allEvents, removeExcessDetails(eventsJSON)...)
	}
	return allEvents
}
