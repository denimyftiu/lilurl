### LIL-URL
URL shortner service with cache support (Redis) and persistent store (Postgres).

#### To run all the infrastructure:

```bash
docker-compose up
```
#### To compile and run only the shortner server:

```bash
go build -o lilurl ./cmd/shortner
./lilurl
```

The environment configuration for connecting to postgres and redis
can be found in: `./pkg/config/config.go`.
