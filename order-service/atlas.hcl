env "local" {
  url = "postgres://devuser:devpass@localhost:5556/order_service?sslmode=disable"
  dev = "postgres://devuser:devpass@localhost:5556/order_service?sslmode=disable"

  schema {
    src = "file://dbschema.hcl"
  }

  migration {
    dir = "file://migrations"
  }
}
