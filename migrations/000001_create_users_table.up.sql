-- Active: 1659338480047@@127.0.0.1@5432@user_db
CREATE TABLE IF NOT EXISTS auth_user(
    id BIGSERIAL PRIMARY KEY,
    firstname TEXT NOT NULL,
    lastname TEXT NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    username TEXT NOT NULL,  
    active BOOLEAN NOT NULL,
    role INTEGER,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1
)
