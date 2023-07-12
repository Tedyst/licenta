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