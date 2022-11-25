FROM node:16.15.1 AS builder
WORKDIR /app
COPY package.json ./
COPY . .
ARG REACT_APP_GOOGLE_MAPS_API_KEY=NO_KEY
ENV REACT_APP_GOOGLE_MAPS_API_KEY=$REACT_APP_GOOGLE_MAPS_API_KEY
RUN echo $REACT_APP_GOOGLE_MAPS_API_KEY
RUN npm install
RUN npm run build

FROM golang:1.18 AS server
WORKDIR /app
COPY --from=builder /app/frontend/build /app/frontend/build

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/app ./

EXPOSE 5000
CMD ["app"]