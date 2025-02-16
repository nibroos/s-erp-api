<h1 align="center">
    <br>
  Backend S-ERP-API
  <br>
</h1>

## 🚀 Quick Start
### Developement Environment
On `/` dir, Run `make copy-env`, Modify to suit your environment, focus on these key, you can leave others as it is. The key name is explanatory itself.
```bash
# MAIN APP PORT
GATEWAY_PORT=4050

# DATABASE
POSTGRES_PORT=5432
```

> Without docker, you need to install [air-verse](https://github.com/air-verse/air) to activate the hot reloading.

### 🐳 Docker :: Container Platform

[Docker](https://docs.docker.com/get-docker/) Install.

- On the root folder, Starts the containers in the background and leaves them running : `docker-compose -f docker/docker-compose-dev.yml up --build -d`
- Stops containers and removes containers, networks, volumes, and images : `docker-compose down`

## 🛎 Available Commands each Service

Change bash directory to each service.
> ${arg} means replace all of it match your args without space
- Run export path : `export PATH="$PATH:$(go env GOPATH)/bin"`
- Create mirgration : `make migrate-create name=${your_migration_name}`
- Run migration : `make migrate-up`
- Stepback migraiton: `make migrate-down`
- Generate proto file, leave the proto args blank if you want to generate all proto file: `make proto ${your-proto.proto}`. If its fail, run this command on specific service. for example, in /service/ run bash `export PATH="$PATH:$(go env GOPATH)/bin"`
- Create seeder : `make seed-create name=${your_seeder_name}`
- Run seeder : `make seed-run file=${your_seeder_name}.sql`

## 💎 The Package Features

<p>
  <img src="https://img.shields.io/badge/-Docker-2496ED?style=for-the-badge&logo=Docker&logoColor=fff" />&nbsp;&nbsp;
  <img src="https://img.shields.io/badge/-NGINX-269539?style=for-the-badge&logo=NGINX&logoColor=fff" />
  <img src="https://img.shields.io/badge/-Go-1185F4?style=for-the-badge&logo=Go&logoColor=fff" />
</p>
<p>
<img src="https://img.shields.io/badge/-PostgreSQL-336791?style=for-the-badge&logo=PostgreSQL&logoColor=fff" />&nbsp;&nbsp;
</p>

## 📔 Notes & Issues

#### dial tcp: lookup postgres: no such host
Change the makefile DB_HOST to `localhost` if run in local env, when running on docker, change it to `postgres`, make sure no space in the value.

#### run multiple seeder in one execution
You can run multiple seeder references in the seeder_controller.go file with password on body payload = env of POSTGRES_PASSWORD.

#### error running migration fix migration
Change the 'version' column name on schema_migrations to latest succeed migration, change the 'dirty' column to false, then run the migration again

#### error function gen_salt(unknown) does not exist, postgre extensions
`CREATE EXTENSION IF NOT EXISTS pgcrypto;`

### 📗 API Document
All endpoints stored in `S-ERP-API.postman_collection.json`.
Postman credentials:
- Email: `yubi@email.com`
- Password: `erppostman1!`

### 🖥️ Prometheus & Grafana
- Grafana will fail at first run, because user credentials need to be created manually in the postgres database.
- Login with the user and password. The default user and password is `admin` and `admin`.
- Dashboard ID
  - PostgreSQL: `9628`
  - Node Exporter: `1860`
  - HTTP Request: `S-ERP-API.postman_collection.json`
- Use `vegeta` to generate the load test. look up `target.txt` for the target URL, then run `vegeta attack -targets=target.txt -rate=1000 -duration=30s -output /dev/null`, change the rate and duration as you need.
- Use `k6` to execute the test. look up `load-test` folder for the script, then run `k6 run script.js`, change the script as you need.


<h1 align="center">
    <br>
  Features
  <br>
</h1>

Feel free to ask if you have any questions or need more details!

