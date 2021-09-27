# Image Repo

## Overview

Image Repo is a self-contained image repository that contains the following features:

- ADD image(s) to the repository
    - private or public (permissions)
    - secure uploading and stored images
- DELETE image(s)
    - prevent a user deleting images from another user (access control)
    - secure deletion of images

## Development

### Required Dependecies:
- Docker
- Git

### Development Setup

1. Clone the project

```bash
git clone https://github.com/adithya/image-repo
```

Before running the application an .env file needs to be created at the root of the project directory:

2. This file will contain the following:
   1. Credentials for a postgressql user
   2. An HS256 JWT secret
   3. An IS_DEBUG atttribute set to "true"
   4. A POSTGRES_DB attribute set to "shopify-challenge-db"
   5. A PGADMIN_LIST_PORT attribute set to "5432"
   6. A CLOUD_STORAGE_HOST attribute set to "localhost"

Here is a sample of how the .env file should look:
```
POSTGRES_HOST=localhost
POSTGRES_USER=postgres
POSTGRES_PASSWORD=secret
JWT_SECRET=UwaLXj%nGl:wR0f4]:F1[H;f(}5ent/Zit{Nc7SCnhg%aZpl9qdoqlFH}Q}(5kG
IS_DEBUG=true
POSTGRES_DB=shopify-challenge-db
PGADMIN_LISTEN_PORT=5432
CLOUD_STORAGE_HOST=localhost
```

You can generate a JWT secret [here](https://www.grc.com/passwords.htm).

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
- A locally running [emulator of Google Cloud Storage](https://github.com/fsouza/fake-gcs-server) will be running on http://localhost:4443

## API Docs

API documentation for image-repo can be found in the [repositories wiki](https://github.com/adithya/image-repo/wiki/API-Reference-Home).

## Secure storage and retrieval of images

### Overview

In an image repository with both public and private access, as well as support for an arbitrary number of users, secure storage and retrieval of images is paramount. This is supported in image-repo with the following techniques:

- Image(s) are stored and retrieved based on ACLs maintained by the back-end, and stored in the DB
- To provide a further layer of security, there is a designated bucket for public images, and each user has their own bucket for their private images
- When the visibility of an image is changed, images are moved to either the users private bucket, or to the public bucket depending on what the new visibility setting is
- [Signed URLs](https://cloud.google.com/storage/docs/access-control/signed-urls) are used for all images with a five hour expiry on the URL

### Next Steps

- While Signed URLs with a timed expiry is good, in reality this technique has drawback, access to the URL is abritrary. If the URL is stolen or shared, and the URL is still within its expiry window, access to a private image could occur. Obviously this is less than ideal, in order to solve this problem a CDN like Google Cloud CDN needs to be used, as it supports [signed URLs and signed cookies](https://cloud.google.com/cdn/docs/private-content), ensuring only those clients that have the signed cookie (in our case the specific user) can have access to the content.



