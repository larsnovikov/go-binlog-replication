# Mysql binlog replicator

##### Based on [JackShadow/go-binlog-example](https://github.com/JackShadow/go-binlog-example) 

### Site 

See full docs [here](https://larsnovikov.github.io/horgh) 

### Master types
- MySQL

### Slave types
- MySQL
- PostgreSQL
- Yandex ClickHouse
- HP Vertica

### Quick start

See quick start tutorial [here](https://larsnovikov.github.io/horgh#quick_start) 

### Testing
- Copy `src/.env.dist` to `src/.env` and set credentials.
- Configure your my MySQL master as `examples/master/mysql.conf`. 
Don't forget to set `binlog_do_db=<master_db_name>` and restart MySQL service.
- Create databases and tables as `examples/sql/`.
- Start Docker as `make start-dev`
- Run as `cd src` and `go run main.go listen` in docker container.
- Copy `examples/configs/user.json` and `examples/configs/post.json` to `src/system/configs`
- Execute `cd src` and `go run main.go load`

### Add tables to replicator

- Use `create-model <table>` to create json config for your table.
- Create table on slave.
- Set `<table>` to `TABLE` param in `src/.env` 
- Use `build-slave` to copy table data from master and set start position of log for table listener.

### Custom handlers for field value

- Create `plugins/user/<plugin_name>/handler.go` like `plugins/system/set_value/handler.go`
- Execute `go build -buildmode=plugin -o plugins/user/<plugin_name>/handler.so plugins/user/<plugin_name>/handler.go`
- Add to field description in your `src/system/configs/<model>.json`

```
"beforeSave": {
  "handler": "user/<plugin_name>",
  "params": [
    "***"
  ]
}
```

##### System handlers

- If you want to set custom field value use `system/set_value` as `handler` param. Don't forget to set `params: ["<value>"]`

### Tools

- `set-position <table> <binlog_name> <binlog_position>` set start position of log for table listener
- `load` start loader for replication testing (for default tables user and post)
- `create-model <table>` create model json-file by master table structure
- `build-slave` create master table dump, restore this dump in slave, set start position of log for table listener
- `destroy-slave` truncate table, set empty position of log for table listener

### Modes

- Prod mode: build app and execute listener. 
  
  Use `make build-prod` and `make start-prod` to start and `make stop-prod` to stop.
- Dev mode: provides the opportunity for manual start and debug. 
  
  Use `make start-dev` to start and `make stop-dev` to stop.