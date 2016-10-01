# Running mysql command in docker container

According to the `docker-compose.yml`, this directory is mounted as `/sql`.

To execute the sqls to a running container, do the following.

```
docker exec webapp_mysql_1 sh -c 'MYSQL_PWD=password mysql -uroot < /sql/create_database.sql'
docker exec webapp_mysql_1 sh -c 'MYSQL_PWD=password mysql -uroot -Disuketch < /sql/schema.sql'
zcat /sql/initial_data.sql.gz | MYSQL_PWD=password mysql -uroot -Disuketch --default-character-set=utf8mb4
```
