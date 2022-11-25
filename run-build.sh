docker build --build-arg REACT_APP_GOOGLE_MAPS_API_KEY={KEY} -t curated-concerts .
docker run -p 5000:5000 --name curated-concerts curated-concerts