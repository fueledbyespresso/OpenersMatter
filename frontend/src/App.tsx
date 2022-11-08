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
        console.log(process.env)
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
                    <ReactGoogleAutocomplete
                        apiKey={process.env.REACT_APP_GOOGLE_MAPS_API_KEY}
                        onPlaceSelected={(place: any) => {
                            if(place != undefined){
                                setPlace(place)
                            }
                        }}
                    />
                    <label>
                        Radius
                        <input placeholder={"Radius"}/>
                    </label>
                    {loadingConcerts && <div>Loading Concerts</div>}

                    {concerts && (
                        <div className="concerts">
                            {Object.keys(concerts).map((key, val) => {
                                return (
                                    concerts[key] && (
                                        <div key={key} className={"concert-card"}>
                                            <img src={concerts[key].images[0].url}/>
                                            <div>{concerts[key].name}</div>
                                            <div>{concerts[key].startDate}</div>
                                            {Object.keys(concerts[key].attractions).map((attraction) => {
                                                return (
                                                    <div className={`${concerts[key].attractions[attraction] ? "favorite" : "not-favorite"}`}>{attraction}</div>
                                                )
                                            })}
                                            <a href={concerts[key].url}>TicketMaster</a>
                                        </div>
                                    )
                                );
                            })}
                        </div>
                    )}
                    <button onClick={() => getConcerts()}>Find Concerts</button>
                </div>
            ) : (
                <a href={"./oauth/v1/login"}>Login</a>
            )}
        </div>
    );
}

export default App;
