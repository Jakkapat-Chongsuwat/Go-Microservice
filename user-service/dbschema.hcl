schema "public" {}

table "users" {
  schema = schema.public

  column "id" {
    type = varchar(255)
    null = false
  }

  column "username" {
    type = varchar(255)
    null = false
  }

  column "email" {
    type = varchar(255)
    null = false
  }

  column "created_at" {
    type = timestamp
    null = false
    default = sql("CURRENT_TIMESTAMP")
  }

  column "updated_at" {
    type = timestamp
    null = false
    default = sql("CURRENT_TIMESTAMP")
  }

  column "deleted_at" {
    type = timestamp
    null = true
  }

  primary_key {
    columns = [column.id]
  }
}

function "set_updated_at" {
  schema = schema.public
  lang   = PLpgSQL
  return = trigger
  as = <<-SQL
  BEGIN
    NEW.updated_at := now();
    RETURN NEW;
  END;
  SQL
}

trigger "users_update_timestamp_trigger" {
  on = table.users
  before {
    update = true
  }
  execute {
    function = function.set_updated_at
  }
}
