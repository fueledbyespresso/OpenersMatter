import React, {useEffect, useState} from 'react';
import './app.scss';

function App() {
    const [user, setUser] = useState<string | null>(null)
    const [loadingConcerts, setLoadingConcerts] = useState<boolean>(false)
    const [concerts, setConcerts] = useState<any | null>(null)
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
        fetch("/api/v1/concerts")
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
                        Longitude
                        <input placeholder={"Longitude"}/>
                    </label>
                    <label>
                        Latitude
                        <input placeholder={"Latitude"}/>
                    </label>
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
                                            {key}: {concerts[key]}
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
