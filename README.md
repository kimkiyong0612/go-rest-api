# go-rest-api
CRUD application
> This architecture is affected flat-package(https://future-architect.github.io/articles/20201109/)

```
sql-migrate https://github.com/rubenv/sql-migrate
sqlx https://github.com/jmoiron/sqlx
validator https://github.com/go-playground/validator
logrus https://github.com/sirupsen/logrus
godotenv https://github.com/joho/godotenv
```

## Set up
- create volume for docker
```zsh
docker volume create go-rest-api-mysql8_volume
```

## Run APP
```
make run
```
