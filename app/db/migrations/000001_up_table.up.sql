CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- updated_at column trigger
CREATE OR REPLACE FUNCTION set_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS "registered"(
    "id" UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    "dev_eui" varchar(255) UNIQUE NOT NULL,
    "status" BOOLEAN NOT NULL DEFAULT FALSE,
    "updated_at" TIMESTAMP,
    "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()

);

CREATE TABLE IF NOT EXISTS "idempotency"(
    "id" UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    "key" varchar(255) UNIQUE NOT NULL,
    "updated_at" TIMESTAMP,
    "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);




create trigger idempotency_updated_at 
  before update on idempotency
  for each row
  execute procedure set_updated_at_column();


create trigger registered_updated_at 
  before update on registered
  for each row
  execute procedure set_updated_at_column();



