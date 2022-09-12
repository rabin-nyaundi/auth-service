
<h1 align="center">
  <br>
  <a href="http://www.amitmerchant.com/electron-markdownify"><img src="https://raw.githubusercontent.com/amitmerchant1990/electron-markdownify/master/app/img/markdownify.png" alt="Markdownify" width="200"></a>
  <br>
  Auth Service
  <br>
</h1>

<h4 align="center">Complete authentication service built by <a href="https://golang.org" target="_blank">Go</a>, <a href="http://electron.atom.io" target="_blank">Postgres</a>, database, stateful token authentication.</h4>

<br />

<p align="center">
  <a href="#key-features">Key Features</a> •
  <a href="#how-to-use">How To Use</a> •
  <a href="#download">Download</a> •
  <a href="#credits">Credits</a> •
  <a href="#related">Related</a> •
  <a href="#license">License</a>
</p>

![screenshot](https://raw.githubusercontent.com/amitmerchant1990/electron-markdownify/master/app/img/markdownify.gif)

## Key Features

* User registration.
  - Register a user with firstname, lastname, email, username and password.
* Account activation
  - After reistration, activation link sent to email with activation tiken appended to it.
* Token authentication
  - Sending a post request with correct email address and password return a token.
* Retrive user and all users
  - Protected endpoint that retrieves all users and a specifc user provded with a token.
* Email notification
  - Email notifcation sent after registrtion.

# How To Use

To clone and run this application, you'll need [Go](https://golang.org) , [Postgres database](https://www.postgresql.org/), [make](https://www.gnu.org/software/make/) and [docker](https://www.docker.com/) installed on your computer. From your command line:


## Run without docker

```bash
# Clone this repository
$ git clone https://github.com/amitmerchant1990/electron-markdownify

# Go into the repository
$ cd auth-service

# Create .envrc file
$ touch .envrc

  Add the following to `.envrc` file

  export DATABASE_DSN=postgresql://<database_user>:<password>@localhost/<database_name>?sslmode=disable

# Create .env file
$ touch .env

  Add the following to `.env` file

    SMTP_HOST=your_smtp_host
    SMTP_PORT=syour_mtp_port
    SMTP_USERNAME=your_smtp_username
    SMTP_PASSWORD=your_smtp_password
    SMTP_SENDER=Task<no-reply@auth-service.com>

# Apply database migrations
$ make db/migrations/new

# Run the app
$ make run/api

# Visit the application url
Visit localhost:4002
```

## Run with docker

```bash
# Clone this repository
$ git clone https://github.com/amitmerchant1990/electron-markdownify

# Go into the repository
$ cd auth-service

# Create .env file
$ touch .env

  Add the following to `.env` file

    DATABASE_DSN=postgresql://<database_user>:<password>@localhost/<database_name>?sslmode=disable

    SMTP_HOST=smtp_host
    SMTP_PORT=smtp_port
    SMTP_USERNAME=smtp_username
    SMTP_PASSWORD=smtp_password
    SMTP_SENDER=Task<no-reply@auth-service.com>

# Run docker compose up
$ docker-compose up

# Visit the application url
Visit localhost:4002
```
# API endpoints

## REST API example application

The REST API to the example app is described below.

### Register user

### Request

`POST /v1/users/`

    BODY='{"firstname":"test", "lastname":"user", "username":"testuser", "email":"test@test.com", "password": "pass@55word"  }'
    
    curl -X POST -d "$BODY" http://localhost:4002/v1/users/

### Response

    {
      "user": "user created successfuly"
    }


## Activate user account

### Request

`POST /v1/users/activated/`

    BODY='{"token": "HSKBAPCPVB5I7P627SOH2OKOPA"}', 

    curl -X POST -d "$BODY" http://localhost:4002/v1/users/activated

### Response

    {
      "data": {
        "id": 2,
        "firstname": "test",
        "lastname": "user",
        "username": "testuser",
        "email": "test@test.com",
        "active": true,
        "role": 1,
        "CreatedAt": "2022-09-12T20:43:57Z",
        "UpdatedAt": "2022-09-12T20:44:21.050272223Z"
      }
    }

## Get authentication token

### Request

`POST /v1/token/authenticate`

    BODY='{ "email":"rabin.nyaundi254@gmail.com", "password": "pass@55word"}'

    curl -X POST -d "$BODY" http://localhost:4002/v1/token/authenticate

### Response

    {
      "authentication_token": {
        "token": "FCK7CD6KHJCV2UDPONJVAXFWGQ",
        "expiry": "2022-09-13T20:52:39.605213672Z"
      }
    }


## Credits

This software uses the following open source packages:

- [Go ](http://go.org/)
- [PostgreSQL](https://www.postgresql.org)
- [Docker](https://www.docker.com/)
- [Make](https://www.gnu.org/software/make/)