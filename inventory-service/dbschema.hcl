schema "public" {}

table "public" "products" {
  schema = schema.public

  column "id" {
    type    = varchar(255)
    null    = false
    default = sql("gen_random_uuid()")
  }

  column "name" {
    type = varchar(255)
    null = false
  }

  column "quantity" {
    type = int
    null = false
  }

  column "price" {
    type = numeric
    null = false
  }

  column "created_at" {
    type    = timestamp
    null    = false
    default = sql("CURRENT_TIMESTAMP")
  }

  column "updated_at" {
    type    = timestamp
    null    = false
    default = sql("CURRENT_TIMESTAMP")
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
      IF NEW.updated_at IS NULL THEN
        NEW.updated_at := timezone('utc', now());
      END IF;
      RETURN NEW;
    END;
  SQL
}

trigger "products_update_timestamp_trigger" {
  on = table.public.products
  before {
    update = true
  }
  execute {
    function = function.set_updated_at
  }
}

# atlas migrate apply --dir file://migrations --env local --baseline 20250214180345 --revisions-schema atlas_schema_revisions
# atlas migrate diff create_orders_tables --env local
# atlas migrate hash
# atlas migrate status --dir file://migrations --env local --revisions-schema atlas_schema_revisions
# atlas migrate apply --env local --revisions-schema atlas_schema_revisions