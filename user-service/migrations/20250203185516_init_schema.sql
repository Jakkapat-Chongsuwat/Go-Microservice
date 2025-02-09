-- Add new schema named "public"
CREATE SCHEMA IF NOT EXISTS "public";
-- Create "set_updated_at" function
CREATE FUNCTION "public"."set_updated_at" () RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  NEW.updated_at := now();
  RETURN NEW;
END;
$$;
-- Create "users" table
CREATE TABLE "public"."users" ("id" character varying(255) NOT NULL, "username" character varying(255) NOT NULL, "email" character varying(255) NOT NULL, "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, "deleted_at" timestamp NULL, PRIMARY KEY ("id"));
-- Create trigger "users_update_timestamp_trigger"
CREATE TRIGGER "users_update_timestamp_trigger" BEFORE UPDATE ON "public"."users" FOR EACH STATEMENT EXECUTE FUNCTION "public"."set_updated_at"();
