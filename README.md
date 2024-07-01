# Go API Server

## Description
Backend Go API application that can be used in production.[^1] The application can add, remove, list, and update customers that are saved in a PostgreSQL database. It uses a redis container as a caching layer. The project has examples of: 

&emsp;*Database interactions.* \
&emsp;*Data serialization/deserialization.* \
&emsp;*Full error handling.* \
&emsp;*Unit testing.* \
&emsp;*Pagination for "all" retrieval.* \
&emsp;*Redis caching.* \
&emsp;*Containerization orchestration.* \
&emsp;*Full Rest API CRUD functionalities.*

Future examples to be included: 

&emsp;*API gateway/security.* \
&emsp;*Goroutine/multithreading.* 


[^1]: The .env file is not encrypted and all secrets are visible for ease of use. If you intend to use this for production, please encrypt the .env file and change the variable values!

## Motivation
Go is widely used for backend programming. It is essential to know how to create a backend API server with basic functionalities. I wanted to take it further by making the project as robust and production ready as possible.

## Quick Start

You don't need Go installed on your computer, but you must have Docker up and running.

Clone the repo:

```
git clone https://github.com/capgainschristian/go_api_ds.git
```

Start the application:

```
docker compose up
```

## Usage

### Adding customers
After you have the application up and running, you will notice that you have no customers to view. I have created a function to generate 100 random customers. To run it, open another terminal and do the following:

```
docker exec -it go_api_ds-web-1 bash
```
Once you are inside the container, run:

```
go run fake_data_generation/fake_data.go
```

### View customers
To verify that the customers have been added successfully, you can run:

```
curl http://localhost:3000/listcustomers
```

Or open a browser and go to:

```
http://localhost:3000/listcustomers
```

Since pagination is used to make data retrieval more efficient and user friendly, you will only see 10 customers listed per page by default. You can change this by appending:

```
listcustomers?limit=100&offset=0
```

Example:

```
http://localhost:3000/listcustomers?limit=100&offset=0
```

### Run unit tests
If you want to run the unit tests, run the following inside the *go_api_ds-web-1* container:

```
cd tests

go test -v
```
### Adding, deleting, and updating a user manually

Make sure that the application is running. Open another terminal and follow the instructions below. If a user with the same information is already created, the examples will not work. In that situation, change the customer information.

**NOTE:** You must be authenticated to add, update, or delete customers. Once you signup for an account, you will receive a token for authentication. The token is added to your cookie automatically. Therefore, it is easier to run these APIs with Postman or VSCode Thunder Client. Otherwise, you will need to include your token in all of your curl requests.

To signup for an account:

```
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"email":"alexander.graham@grahamsummitllc.com,","password":"thisisaverystrongpassword"}' \
  http://localhost:3000/signup
```

To login/auithenticate with your new account:

```
curl --header "Content-Type: application/json" \
  --verbose \
  --request POST \
  --data '{"email":"alexander.graham@grahamsummitllc.com,","password":"thisisaverystrongpassword"}' \
  http://localhost:3000/login
```

To add a customer:

```
curl --header "Content-Type: application/json" \
  --request POST \
  -b "token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MjI0NTI3OTMsInN1YiI6ImFsZXhhbmRlci5ncmFoYW1AZ3JhaGFtc3VtbWl0bGxjLmNvbSwifQ.tO7v42pkJHqeX81g4yG2apuRGv1YGtGpN9Wrmre4NBg" \
  --data '{"name":"Christian Graham,","email":"christian.graham@grahamsummitllc.com","address":"777 Summit LLC Drive","number":2222}' \
  http://localhost:3000/addcustomer
```

To update a customer:

```
curl -X PUT http://localhost:3000/updatecustomer \
      -b "token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MjI0NTI3OTMsInN1YiI6ImFsZXhhbmRlci5ncmFoYW1AZ3JhaGFtc3VtbWl0bGxjLmNvbSwifQ.tO7v42pkJHqeX81g4yG2apuRGv1YGtGpN9Wrmre4NBg" \
     -H "Content-Type: application/json" \
     -d '{
           "email": "christian.graham@grahamsummitllc.com",
           "name": "Christian Graham",
           "address": "888 Summit LLC Drive",
		   "number": 1111
         }'
```

To delete a customer:

```
curl --header "Content-Type: application/json" \
  --request DELETE \
  -b "token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MjI0NTI3OTMsInN1YiI6ImFsZXhhbmRlci5ncmFoYW1AZ3JhaGFtc3VtbWl0bGxjLmNvbSwifQ.tO7v42pkJHqeX81g4yG2apuRGv1YGtGpN9Wrmre4NBg" \
  --data '{"email":"christian.graham@grahamsummitllc.com"}' \
  http://localhost:3000/deletecustomer
```
### PostgreSQL

If you ever need to get into the PostgreSQL container, run the following:

```
docker exec -it go_api_ds-db-1 psql -U capgainschristian -d customers
```
Once you're inside, you can run queries normally like:

```
SELECT * FROM customers;
```

### Working on the application

As stated above, you don't need Go installed but you do need Docker to run the application. The same can be said when working on the project. If you want to further develop, or edit, any of Go code, simply run:

```
docker compose run --service-ports web bash
```
This will take you into a container with Go already installed. After making your changes, you can run the app from the container:

```
go run cmd/main.go -b 0.0.0.0
```
