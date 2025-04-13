-- Create "set_updated_at" function
CREATE OR REPLACE FUNCTION public.set_updated_at() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  NEW.updated_at := CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$;

-- Create "orders" table
CREATE TABLE public.orders (
  "id" character varying(255) NOT NULL DEFAULT gen_random_uuid(),
  "user_id" character varying(255) NOT NULL,
  "status" character varying(255) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamp NULL,
  PRIMARY KEY ("id")
);

-- Create "order_items" table
CREATE TABLE public.order_items (
  "id" character varying(255) NOT NULL DEFAULT gen_random_uuid(),
  "order_id" character varying(255) NOT NULL,
  "product_id" character varying(255) NOT NULL,
  "quantity" integer NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamp NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_order_items_order" FOREIGN KEY ("order_id") REFERENCES public.orders ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);

-- Create index "idx_order_items_order_id" to table: "order_items"
CREATE INDEX "idx_order_items_order_id" ON public.order_items ("order_id");

-- Create trigger "order_items_update_timestamp_trigger"
CREATE TRIGGER "order_items_update_timestamp_trigger" 
BEFORE UPDATE ON public.order_items 
FOR EACH STATEMENT 
EXECUTE FUNCTION public.set_updated_at();

-- Create trigger "orders_update_timestamp_trigger"
CREATE TRIGGER "orders_update_timestamp_trigger" 
BEFORE UPDATE ON public.orders 
FOR EACH STATEMENT 
EXECUTE FUNCTION public.set_updated_at();