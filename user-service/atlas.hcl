env "local" {
  url = "postgres://devuser:devpass@localhost:5555/user_service?sslmode=disable"
  dev = "postgres://devuser:devpass@localhost:5555/user_service?sslmode=disable"

  schema {
    src = "file://dbschema.hcl"
  }

  migration {
    dir = "file://migrations"
  }
}
