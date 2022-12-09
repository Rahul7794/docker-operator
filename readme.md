## docker-operator

### intro

This is the rest API that accepts request to run and launch docker container and print the response. The project uses Fiber framework https://docs.gofiber.io/

### install

- first install Go
- clone & navigate into the root of the project 
- go run server.go (to run the api)
- make test (to run tests)

### environment

Set up below environment variables before starting the application locally.

```
API_PORT=:8080                    // API_PORT to be exposed
REGISTRY=docker registery         // Docker registery URL
LOG_LEVEL=DEBUG                   // Set Log Level for application
LOG_PATH=logs/                    // path of file where logs would be written to (only works with LOG_WRITE_MODE=file) 
LOG_WRITE_MODE=file               // log write mode (console/file)
CONTENT_LENGTH=10                 // lenght of content to be logged
```

### endpoints

#### Endpoint /api/status :<br />
GET: healthCheck -> To check the health of application

#### Endpoint /api/exec/:image_name/:tag :<br />
GET: exec -> To run the docker image with a tag passed <br />
Response: content returned by docker

#### Endpoint /api/exec/:image_name/:tag :<br />
POST: exec -> To run the docker image with a tag passed <br />
Response: content returned by docker

### examples

#### POST

Runs an image with the params required by docker image to run. 

```
POST /api/exec/infosicav/latest
Host: localhost:8080
Content-Type: application/json

{"data": ["00000X71080", "json"]}

or

curl -i --header "Content-Type: application/json" --request POST --data '{"data": ["00000X71080", "json"]}' http://localhost:8080/api/exec/infosicav/latest
```

if no param is to be passed, we still need to send data with empty request body

```
curl -i --request POST --data '{"data": []}' http://localhost:8080/api/exec/hello_gitlab/20210603-084901
```


#### GET

```
GET /api/exec/hello_world/latest
Host: localhost:8080

or 

http://localhost:8080/api/exec/hello_world/latest
```