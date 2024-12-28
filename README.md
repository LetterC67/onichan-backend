# onichan

backend written in go (gin gonic) for onichan web forum.

the backend use postgresql as database.


## installation
clone the repository:

```
git clone https://github.com/LetterC67/onichan-backend.git
```

navigate to the project directory:
```
cd onichan-backend
```

download the dependencies:
```
RUN go mod download
```

build the project:

```
go build -o main
```

build the additional script too:
```
go build -o script ./scripts
```

## configuration
firstly, create a copy of the `.env.example` file:
```
cp .env.example .env
```

you can then edit the config in `.env` file. you have to config database credentials and jwt secret, which are not preconfigured.

## usage
before running the application, please run the script to migrate the database. this will also create an admin account with username `admin` and password `@dmin123`, which can be changed later. this step only needs to be performed once.

```
./script auto
```


run the application:
```
./main
```

## docker setup
alternatively, one can run the services as docker containers. please config the `.env` file before running:

```
docker compose up -d --build
```

## api docs
access api docs at: `/swagger/index.html`.
