CREATE TABLE 'User'(
    'id' uuid NOT NULL DEFAULT uuid_generate_v4(),
    'username' varchar(255) NOT NULL,
    'password' varchar(255) NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT (now())
)