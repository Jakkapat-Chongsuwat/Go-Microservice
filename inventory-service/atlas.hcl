variable "DB_DRIVER" {
  type    = string
  default = getenv("DB_DRIVER")
}

variable "DB_USER" {
  type    = string
  default = getenv("DB_USER")
}

variable "DB_PASS" {
  type    = string
  default = getenv("DB_PASS")
}

variable "DB_HOST" {
  type    = string
  default = getenv("DB_HOST")
}

variable "DB_PORT" {
  type    = string
  default = getenv("DB_PORT")
}

variable "DB_NAME" {
  type    = string
  default = getenv("DB_NAME")
}

variable "DB_SSLMODE" {
  type    = string
  default = getenv("DB_SSLMODE")
}

variable "DB_SCHEMA" {
  type    = string
  default = getenv("DB_SCHEMA")
}

env "local" {
  url = "${var.DB_DRIVER}://${var.DB_USER}:${var.DB_PASS}@${var.DB_HOST}:${var.DB_PORT}/${var.DB_NAME}?sslmode=${var.DB_SSLMODE}&search_path=${var.DB_SCHEMA}"
  dev = "${var.DB_DRIVER}://${var.DB_USER}:${var.DB_PASS}@${var.DB_HOST}:${var.DB_PORT}/${var.DB_NAME}?sslmode=${var.DB_SSLMODE}&search_path=${var.DB_SCHEMA}"

  schema {
    src = "file://dbschema.hcl"
  }

  migration {
    dir     = "file://migrations"
  }
}
