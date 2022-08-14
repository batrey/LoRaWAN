CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS "user"(
    "id" UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    "username" varchar(255) UNIQUE NOT NULL,
    "password" varchar(255) NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "registered"(
    "id" UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    "deveui" varchar(255) UNIQUE NOT NULL,
    "status" BOOLEAN NOT NULL DEFAULT FALSE,
    "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "idempotency"(
    "id" UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    "idempotencykey" varchar(255) UNIQUE NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
