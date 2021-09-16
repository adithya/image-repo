# Image Repo

## Development

### Required Dependecies:
- Docker
- Google Cloud Platform Account

### Setup

Before running the application two files need to be setup:

1. First a .env file needs to be created, this file will contain the following:
   1. Credentials for a postgressql user
   2. An HS256 JWT secret

The file will look something like this:
```
POSTGRES_USER={username}
POSTGRES_PASSWORD={password}
JWT_SECRET={secret}
```

You can generate a JWT secret [here](https://www.grc.com/passwords.htm).

2. A Google Cloud Platform service account needs to be created, and the JSON key file for that service account needs to be downloaded. Instructions on how do do that can be found [here](https://cloud.google.com/docs/authentication/production#create_service_account).

Both files need to be located at the root of the project directory.

### Getting Started

In order to get up and running with the image-repo, run docker-compose from the root of the project directory:

```bash
docker-compose up
```

Optionally run docker-compose with the "dev" profile for access to a pgadmin instance that is already setup to point to the postgresql server:

```bash
docker-compose --profile dev up
```

By default the services can be found at the following URLs:
- The web server will be available on http://localhost:8080
- pgadmin will be available on http://localhost:80
- postgres will be available on http://localhost:5432