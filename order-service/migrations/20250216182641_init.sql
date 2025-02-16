-- Add new schema named "order_service"
CREATE SCHEMA "order_service";
-- Create "set_updated_at" function
CREATE FUNCTION "order_service"."set_updated_at" () RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  NEW.updated_at := CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$;
-- Create "orders" table
CREATE TABLE "order_service"."orders" ("id" character varying(255) NOT NULL DEFAULT gen_random_uuid(), "user_id" character varying(255) NOT NULL, "status" character varying(255) NOT NULL, "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, "deleted_at" timestamp NULL, PRIMARY KEY ("id"));
-- Create "order_items" table
CREATE TABLE "order_service"."order_items" ("id" character varying(255) NOT NULL DEFAULT gen_random_uuid(), "order_id" character varying(255) NOT NULL, "product_id" character varying(255) NOT NULL, "quantity" integer NOT NULL, "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, "deleted_at" timestamp NULL, PRIMARY KEY ("id"), CONSTRAINT "fk_order_items_order" FOREIGN KEY ("order_id") REFERENCES "order_service"."orders" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- Create index "idx_order_items_order_id" to table: "order_items"
CREATE INDEX "idx_order_items_order_id" ON "order_service"."order_items" ("order_id");
-- Create trigger "order_items_update_timestamp_trigger"
CREATE TRIGGER "order_items_update_timestamp_trigger" BEFORE UPDATE ON "order_service"."order_items" FOR EACH STATEMENT EXECUTE FUNCTION "order_service"."set_updated_at"();
-- Create trigger "orders_update_timestamp_trigger"
CREATE TRIGGER "orders_update_timestamp_trigger" BEFORE UPDATE ON "order_service"."orders" FOR EACH STATEMENT EXECUTE FUNCTION "order_service"."set_updated_at"();
