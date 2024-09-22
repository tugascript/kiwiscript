# Kiwiscript

An online video sharing platform for teaching and learning coding and programming languages.

## Structure

- **Back-end**: go with Fiber and SQLC.
- **Front-end**: rust with leptos and Actix-Web.

### Back-end

The back-end is written in Go using the Fiber framework, and follows a MSC (Model-Service-Controller) pattern. The database is managed using SQLC.

The API is RESTful and follows the [expired Hypertext Application Language (HAL) Internet-Draft](https://datatracker.ietf.org/doc/draft-kelly-json-hal/).

### File Structure

Our project is divided into go packages with the following structure:

```md
├─── app
│      app.go
│      config.go
│      logger.go
│      validator.go
├─── controllers
│      controllers.go
│      helpers.go
│      middleware.go
│      *.controller.go
├─── dtos
│      dtos.go
│      bodies.go
│      query_params.go
│      responses.go
│      *.dto.go
├─── exceptions
│      controller_exceptions.go
│      service_exceptions.go
├─── paths
│      paths.go
├─── providers
│      ├─── cache
│      │      cache.go
│      │      *.go
│      ├─── database
│      │      ├─── migrations
│      │      │      *.down.sql
│      │      │      *.up.sql
│      │      ├─── queries
│      │      │      *.sql
│      │      database.go
│      │      db.go
│      │      models.go
│      │      *.model.go
│      │      *.sql.go
│      ├─── email
│      │      email.go
│      │      *.go
│      ├─── oauth
│      │      oauth.go
│      │      *.go
│      ├─── object_storage
│      │      object_storage.go
│      │      *.go
│      ├─── tokens
│      │      tokens.go
│      │      *.go
├─── routers
│      routers.go
│      *.router.go
├─── services
│      services.go
│      helpers.go
│      *.service.go
├─── utils
│      utils.go
│      *.go
├─── test
│      ├─── fixtures
│      │      *.pdf/jpeg/...
│      common_test.go
│      *_e2e_test.go
```

### Database Model

TODO

### Local Setup

1. Cd into the `kiwiscript_go` directory.
2. Install Go.
3. Copy the `.env.example` file to `.env` and fill in the required environment variables.
4. Generate the JWT keys using the following command:

    ```bash
    make keygen
    ```

5. Start DB, Cache and S3 services with Docker-Compose or Podman-Compose.

    ```bash
    # For Docker
    make dc-up
    # For Podman
    make pm-up
    ```

6. Create the database.

    ```bash
    # For Docker
    make create-db
    # For Podman
    make create-pm-db
    ```

7. Install go-migrate and run the migrations.

    ```bash
    make migrate-up
    ```

8. Run the server.

    ```bash
    make dev
    ```

### Testing

1. Cd into the `kiwiscript_go` directory.
2. Start DB, Cache and S3 services with Docker-Compose or Podman-Compose.

    ```bash
    # For Docker
    make dc-up
    # For Podman
    make pm-up
    ```

3. Create the test database.

    ```bash
    # For Docker
    make create-test-db
    # For Podman
    make create-pm-test-db
    ```

4. Run migrations for the test database.

    ```bash
    make migrate-test-up
    ```

5. Run the tests.

    ```bash
    make test
    ```

### Deployment

TODO

## Front-end

TODO

## License

The code of this project is licensed under the Gnu General Public License v3.0. You can find the license [here](LICENSE).
