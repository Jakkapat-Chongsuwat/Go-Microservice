-- Add new schema named "user_service"
CREATE SCHEMA "user_service";
-- Create "set_updated_at" function
CREATE FUNCTION "user_service"."set_updated_at" () RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  NEW.updated_at := now();
  RETURN NEW;
END;
$$;
-- Create "users" table
CREATE TABLE "user_service"."users" ("id" character varying(255) NOT NULL, "username" character varying(255) NOT NULL, "email" character varying(255) NOT NULL, "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, "deleted_at" timestamp NULL, PRIMARY KEY ("id"));
-- Create trigger "users_update_timestamp_trigger"
CREATE TRIGGER "users_update_timestamp_trigger" BEFORE UPDATE ON "user_service"."users" FOR EACH STATEMENT EXECUTE FUNCTION "user_service"."set_updated_at"();
