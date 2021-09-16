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
- Google Cloud Platform Account
- Git

### Setup

1. Clone the project

```bash
git clone https://github.com/adithya/image-repo
```

Before running the application two files need to be setup:

2. First a .env file needs to be created, this file will contain the following:
   1. Credentials for a postgressql user
   2. An HS256 JWT secret

The file will look something like this:
```
POSTGRES_USER={username}
POSTGRES_PASSWORD={password}
JWT_SECRET={secret}
```

You can generate a JWT secret [here](https://www.grc.com/passwords.htm).

3. A Google Cloud Platform service account needs to be created, and the JSON key file for that service account needs to be downloaded. Instructions on how do do that can be found [here](https://cloud.google.com/docs/authentication/production#create_service_account).

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

## Secure storage and retrieval of images

### Overview

In an image repository with both public and private access, as well as support for an arbitrary number of users, secure storage and retrieval of images is paramount. This is supported in image-repo with the following techniques:

- Image(s) are stored and retrieved based on ACLs maintained by the back-end, and stored in the DB
- To provide a further layer of security, there is a designated bucket for public images, and each user has their own bucket for their private images
- When the visibility of an image is changed, images are moved to either the users private bucket, or to the public bucket depending on what the new visibility setting is
- [Signed URLs](https://cloud.google.com/storage/docs/access-control/signed-urls) are used for all images with a five hour expiry on the URL

### Next Steps

- While Signed URLs with a timed expiry is good, in reality this technique has drawback, access to the URL is abritrary. If the URL is stolen or shared, and the URL is still within its expiry window, access to a private image could occur. Obviously this is less than ideal, in order to solve this problem a CDN like Google Cloud CDN needs to be used, as it supports [signed URLs and signed cookies](https://cloud.google.com/cdn/docs/private-content), ensuring only those clients that have the signed cookie (in our case the specific user) can have access to the content.

## API Docs

API documentation for image-repo can be found in the repositories wiki.


