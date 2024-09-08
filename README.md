# API Faker

API Faker is a RESTful API built with Go that provides endpoints for managing accounts and projects. The project uses PostgreSQL as the database and features dynamic configuration using environment variables.

## Features

- CRUD operations for Accounts and Projects
- PostgreSQL database connection
- Modular project structure
- Configurable using environment variables
- Simple mock endpoints for testing

## Requirements

- [Go](https://golang.org/doc/install) 1.18 or later
- [PostgreSQL](https://www.postgresql.org/download/) (running locally or on a remote server)

## Setup

### 1. Clone the Repository

Clone the repository to your local machine:

git clone https://github.com/adolfooes/api_faker.git

cd api_faker

### 2. Install Go Modules

Install the Go dependencies required for the project:

go mod tidy

This will ensure that all dependencies specified in `go.mod` are installed.

### 3. Setup PostgreSQL

#### Create a PostgreSQL Database

Make sure you have PostgreSQL installed and running. Create a new PostgreSQL database for the project:

createdb api_faker_db

You can replace `api_faker_db` with the name of your choice.

#### Configure PostgreSQL Connection

Set up the connection to your PostgreSQL database. You will use an environment variable called `DATABASE_URL` to store the connection string.

The format of the connection string is:

postgres://username:password@localhost:5432/dbname?sslmode=disable

Replace `username`, `password`, `localhost`, and `dbname` with your PostgreSQL credentials.

### 4. Set Environment Variable

#### On Linux/macOS

You can export the `DATABASE_URL` as an environment variable in your shell:

export DATABASE_URL="postgres://username:password@localhost:5432/api_faker_db?sslmode=disable"

#### On Windows

Set the environment variable in your terminal:

set DATABASE_URL=postgres://username:password@localhost:5432/api_faker_db?sslmode=disable

Make sure to replace `username`, `password`, and `dbname` with your actual PostgreSQL credentials.

### 5. Run the Project

Once your environment is configured and the database is set up, run the project with the following command:

go run cmd/main.go

You should see the output:

Server is running on port 8080

Database connection established successfully

The server will be running on `http://localhost:8080`.

## API Endpoints

### Accounts

- `GET /accounts`: Retrieve all accounts
- `GET /accounts/{id}`: Retrieve a single account by ID
- `POST /accounts`: Create a new account
- `PUT /accounts/{id}`: Update an account by ID
- `DELETE /accounts/{id}`: Delete an account by ID

### Projects

- `GET /projects`: Retrieve all projects
- `GET /projects/{id}`: Retrieve a single project by ID
- `POST /projects`: Create a new project
- `PUT /projects/{id}`: Update a project by ID
- `DELETE /projects/{id}`: Delete a project by ID

### Mock Data (for testing)

- `GET /mocks`: Retrieve mock data
- `POST /mocks`: Create mock data
- `PUT /mocks`: Update mock data
- `DELETE /mocks`: Delete mock data

## Testing the API

You can use `curl`, Postman, or any API client to test the API.

For example, to retrieve all accounts:

curl -X GET http://localhost:8080/accounts

## Future Improvements

- Add authentication (JWT, OAuth)
- Implement more detailed error handling
- Add more comprehensive unit tests

## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.
