# Proiect de licenta

# Making a migration

```
sql-migrate new <migration_name>
```

After making the migration, edit the file in `db/migrations/<migration_name>.sql` and run the following command to apply the migration:

```
sql-migrate up
```

# Updating the database schema/api

```
go generate ./...
```

# Tools

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install go.uber.org/mock/mockgen@latest
go install github.com/valyala/quicktemplate/qtc@latest
```
