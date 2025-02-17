schema "order_service" {}

table "order_service" "orders" {
  schema = schema.order_service

  column "id" {
    type    = varchar(255)
    null    = false
    default = sql("gen_random_uuid()")
  }

  column "user_id" {
    type = varchar(255)
    null = false
  }

  column "status" {
    type = varchar(255)
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

  column "deleted_at" {
    type = timestamp
    null = true
  }

  primary_key {
    columns = [column.id]
  }
}

table "order_service" "order_items" {
  schema = schema.order_service

  column "id" {
    type    = varchar(255)
    null    = false
    default = sql("gen_random_uuid()")
  }

  column "order_id" {
    type = varchar(255)
    null = false
  }

  column "product_id" {
    type = varchar(255)
    null = false
  }

  column "quantity" {
    type = int
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

  column "deleted_at" {
    type = timestamp
    null = true
  }

  primary_key {
    columns = [column.id]
  }

  index "idx_order_items_order_id" {
    columns = [column.order_id]
  }

  foreign_key "fk_order_items_order" {
    columns     = [column.order_id]
    ref_columns = [table.order_service.orders.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
}

function "set_updated_at" {
  schema = schema.order_service
  lang   = PLpgSQL
  return = trigger
  as = <<-SQL
    BEGIN
      NEW.updated_at := CURRENT_TIMESTAMP;
      RETURN NEW;
    END;
  SQL
}

trigger "orders_update_timestamp_trigger" {
  on = table.order_service.orders
  before {
    update = true
  }
  execute {
    function = function.set_updated_at
  }
}

trigger "order_items_update_timestamp_trigger" {
  on = table.order_service.order_items
  before {
    update = true
  }
  execute {
    function = function.set_updated_at
  }
}

# atlas migrate apply --dir file://migrations --env local --baseline 20250214180345 --revisions-schema atlas_schema_revisions
# atlas migrate diff create_orders_tables --env local
# atlas migrate status --dir file://migrations --env local --revisions-schema atlas_schema_revisions
# atlas migrate apply --env local --revisions-schema atlas_schema_revisions