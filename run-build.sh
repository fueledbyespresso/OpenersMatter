docker build --build-arg REACT_APP_GOOGLE_MAPS_API_KEY=AIzaSyBvtxpvUnt97HL9AIBi7H4AAHFs3cfeJ1k -t curated-concerts .
docker run -p 5000:5000 --name curated-concerts curated-concerts