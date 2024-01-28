<p align="center">
    <img alt="Flagger Logo" src="docs/img/logo512.png" width=400 />
</p>

<p align="center">
    <strong><i>Flagger API</i></strong> is the backend REST API that serves <strong><i>Flagger</i></strong>, the platform that unifies feature flag automation.
</p>

## Installation

### :one: [Install PostgreSQL](https://www.postgresql.org/download/):

Installation process may depend on OS, so **it's best to follow a YouTube tutorial that demonstrates the installation steps.**

### :two: Create Database

Run the following make command to create a new database:

```bash
make create-db
```

### :three: Environment Variables

From the following list of available environment variables, set the ones that are required for local development, as well as any other variables you'd like to customize:

Variable | Default | Description | Required for Local Development?
--- | --- | --- | ---
`FLAGGERAPI_HOSTNAME` | `localhost` | Serving hostname of API | No
`FLAGGERAPI_PORT` | `8080` | Serving port number of API | No
`FLAGGERAPI_SECRET_KEY` | `DEADBEEF` | Secret key used for token encryption | No
`FLAGGERAPI_HASHING_COST` | `14` | Hashing cost for password/key hashing | No
`FLAGGERAPI_FRONTEND_BASE_URL` | `http://localhost:3000` | Frontend URL, used to generate links in emails | No
`FLAGGERAPI_FRONTEND_ACTIVATION_ROUTE` | `/signup/activate/%s` | Frontend activation route, used to generate activation link in emails | No
`FLAGGERAPI_AUTH_ACCESS_LIFETIME` | `300` | Lifetime of access tokens in seconds | No
`FLAGGERAPI_AUTH_REFRESH_LIFETIME` | `2592000` | Lifetime of refresh tokens in seconds | No
`FLAGGERAPI_ACTIVATION_LIFETIME` | `2592000` | Lifetime of activation tokens in seconds | No
`FLAGGERAPI_POSTGRES_HOSTNAME` | `localhost` | PostgreSQL hostname | No
`FLAGGERAPI_POSTGRES_PORT` | `5432` | PostgreSQL port number | No
`FLAGGERAPI_POSTGRES_USERNAME` | `postgres` | PostgreSQL username | **Yes**
`FLAGGERAPI_POSTGRES_PASSWORD` | `postgres` | PostgreSQL password | **Yes**
`FLAGGERAPI_POSTGRES_DATABASE_NAME` | `flaggerdb` | PostgreSQL database name | No
`FLAGGERAPI_SMTP_HOSTNAME` | `smtp.gmail.com` | SMTP hostname | No
`FLAGGERAPI_SMTP_PORT` | `587` | SMTP port number | No
`FLAGGERAPI_SMTP_USERNAME` | `<empty>` | SMTP email username | **Only if `FLAGGERAPI_MAIL_CLIENT_TYPE` is `smtp`**
`FLAGGERAPI_SMTP_PASSWORD` | `<empty>` | SMTP email password | **Only if `FLAGGERAPI_MAIL_CLIENT_TYPE` is `smtp`**
`FLAGGERAPI_MAIL_CLIENT_TYPE` | `console` | Mail client type, must be one of `smtp`, `console`, or `inmem` |  No
`FLAGGERAPI_MAIL_TEMPLATES_DIR` | `templates/email` | Directory to look for email templates | No

### :four: Run Server

Run the server locally using the following make command:

```bash
make server
```

You should see a server log entry indicating that the server has been launched successfully:

```
[I] 2024/01/28 21:36:29 ~/flagger-api/internal/server/controller.go:59: Server running on localhost:8080
```

To terminate the server, use `Ctrl+C`.

## Testing

### Unit Tests

To run all unit tests, run the following make command:

```bash
make test
```
