CREATE TABLE [IF NOT EXISTS] 'user'(
    'id' uuid NOT NULL DEFAULT uuid_generate_v4()PRIMARY KEY,
    'username' varchar(255) UNIQUE NOT NULL,
    'password' varchar(255) NOT NULL,
    'created_at' timestamptz NOT NULL DEFAULT (now())
    'last_login' timestamptz
)

CREATE TABLE [IF NOT EXISTS] 'registered'(
    'id' uuid NOT NULL DEFAULT uuid_generate_v4()PRIMARY KEY,
    'DevEUI' varchar(255) UNIQUE NOT NULL,
    'status' BOOLEAN NOT NULL DEFAULT FALSE,
    'user' varchar(255) NOT NULL,
    'created_at' timestamptz NOT NULL DEFAULT (now())
)