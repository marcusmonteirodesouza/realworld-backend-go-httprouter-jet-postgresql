# ![RealWorld Example App](logo.png)

> ### [Go](https://go.dev/) codebase containing real world examples (CRUD, auth, advanced patterns, etc) that adheres to the [RealWorld](https://github.com/gothinkster/realworld) spec and API.

### [Demo](https://demo.realworld.io/)&nbsp;&nbsp;&nbsp;&nbsp;[RealWorld](https://github.com/gothinkster/realworld)

This codebase was created to demonstrate a fully fledged backend application built with **[Go](https://go.dev/)** including CRUD operations, authentication, routing, pagination, and more.

We've gone to great lengths to adhere to the **[Go](https://go.dev/)** community styleguides & best practices.

For more information on how to this works with other frontends/backends, head over to the [RealWorld](https://github.com/gothinkster/realworld) repo.

# How it works

- It uses a [Modular Monolith architecture](https://www.milanjovanovic.tech/blog/what-is-a-modular-monolith) with clearly defined boundaries and independent modules ([users](internal/services/users.go), [profiles](internal/services/profiles.go), and [articles](internal/services/articles.go)).
- It uses the [Standard Go Project Layout](https://github.com/golang-standards/project-layout) for codebase organization.
- It uses [httprouter](https://github.com/julienschmidt/httprouter) to register routes and handlers. See [routes.go](cmd/api/routes.go).
- It uses [Jet](https://github.com/go-jet/jet) to build type-safe SQL queries. This project doesn't use an ORM.
- It uses [PostgreSQL](https://www.postgresql.org/) as database. Locally, the database runs as container [see docker-compose.yaml](docker-compose.yaml), and all it's data is stored in a volume mapped to the `postgres-data` folder inside the project (the folder is created automatically when running the container).
- It handles errors in a [centralized way](cmd/api/helpers.go).
- It uses [Make](https://www.gnu.org/software/make/) to run some utility scripts. See the [Makefile](Makefile).

# Getting started

1. Run `cp .env.template .env`. The `.env` file contains the environment variables used by both Keycloak and the Realworld backend application, including secrets.

## Run the application

1. Run `make up`.
