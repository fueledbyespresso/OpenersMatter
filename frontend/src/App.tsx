import React, {useEffect, useState} from 'react';
import ReactGoogleAutocomplete from "react-google-autocomplete";
import './app.scss';

function App() {
    const [user, setUser] = useState<string | null>(null)
    const [loadingConcerts, setLoadingConcerts] = useState<boolean>(false)
    const [concerts, setConcerts] = useState<any | null>(null)
    const [curPlace, setPlace] = useState<any | null>(null)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        // Refresh session every 10 minutes
        setInterval(refreshSession, 600000)
        // Listen to localstorage changes
        // Update user on change
        window.addEventListener('storage', () => {
            setUser(JSON.parse(localStorage.getItem("user") || ""))
        });

        //Fetch user from api
        fetch("/oauth/v1/account")
            .then((res) => {
                if (res.ok) {
                    return res.json()
                }
            })
            .then(
                (result) => {
                    setUser(result);
                    localStorage.setItem("user", JSON.stringify(result))
                }, (error) => {
                    setUser(null)
                    localStorage.removeItem("user")
                    setError(error);
                }
            )
    }, [])

    function refreshSession() {
        fetch("/oauth/v1/refresh")
            .then((res) => {
                if (!res.ok) {
                    setError(`Unable to refresh session`)
                }
            })
    }

    function getConcerts() {
        setLoadingConcerts(true)
        fetch("/api/v1/concerts?long="+curPlace?.geometry.location.lng()+"&lat="+curPlace?.geometry.location.lat())
            .then((res) => {
                if (res.ok) {
                    setLoadingConcerts(false)
                    return res.json()
                }
            })
            .then(
                (result) => {
                    setConcerts(result)
                }, (error) => {
                    setConcerts(null)
                    setError(error);
                }
            )
    }

    return (
        <div className="App">
            {user ? (
                <div className={"main"}>
                    <label>
                        Location
                        <ReactGoogleAutocomplete
                            apiKey={process.env.REACT_APP_GOOGLE_MAPS_API_KEY}
                            onPlaceSelected={(place: any) => {
                                if(place != undefined){
                                    setPlace(place)
                                }
                            }}
                        />
                    </label>
                    {loadingConcerts && <div><div className="loader"></div>Loading Concerts</div>}
                    <button onClick={() => getConcerts()}>Find Concerts</button>

                    {concerts && (
                        <div className="concerts">
                            {Object.keys(concerts).map((key, val) => {
                                return (
                                    concerts[key] && (
                                        <div key={key} className={"concert-card"}>
                                            <img src={concerts[key].images[0].url}/>
                                            <div className={"tour-name"}>{concerts[key].name}</div>
                                            <div>{concerts[key].startDate}</div>
                                            <div>
                                                {Object.keys(concerts[key].attractions).map((attraction) => {
                                                    return (
                                                        <div className={`${concerts[key].attractions[attraction] ? "favorite" : "not-favorite"}`}>{attraction}</div>
                                                    )
                                                })}
                                            </div>
                                            <a className="ticketmaster-link" href={concerts[key].url}>TicketMaster</a>
                                        </div>
                                    )
                                );
                            })}
                        </div>
                    )}
                </div>
            ) : (
                <div className="not-logged-in">
                    <h1>Curated Concerts</h1>
                    <h2>
                        Find concerts near you based on your Spotify history!
                    </h2>
                    <label className="spotify-login-button">
                        <a href={"./oauth/v1/login"}>Login with Spotify</a>
                    </label>
                    <img src={"/frontpage-screenshot.png"}/>
                </div>
            )}
        </div>
    );
}

export default App;
