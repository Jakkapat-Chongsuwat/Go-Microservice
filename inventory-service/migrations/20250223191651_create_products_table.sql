-- Create "products" table
CREATE TABLE "products" ("id" character varying(255) NOT NULL DEFAULT gen_random_uuid(), "name" character varying(255) NOT NULL, "quantity" integer NOT NULL, "price" numeric NOT NULL, "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY ("id"));
-- Create "set_updated_at" function
CREATE FUNCTION "set_updated_at" () RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  NEW.updated_at := CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$;
-- Create trigger "products_update_timestamp_trigger"
CREATE TRIGGER "products_update_timestamp_trigger" BEFORE UPDATE ON "products" FOR EACH STATEMENT EXECUTE FUNCTION "set_updated_at"();
