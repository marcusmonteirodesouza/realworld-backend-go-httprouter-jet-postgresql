CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users(
     id UUID CONSTRAINT users_pk PRIMARY KEY DEFAULT gen_random_uuid(),
     email TEXT CONSTRAINT users_email_uk UNIQUE NOT NULL,
     username TEXT CONSTRAINT users_username_uk NOT NULL,
     password_hash TEXT CONSTRAINT users_password_hash_nn NOT NULL,
     bio TEXT,
     image TEXT,
     created_at TIMESTAMP WITH TIME ZONE CONSTRAINT users_created_at_df DEFAULT CURRENT_TIMESTAMP,
     updated_at TIMESTAMP WITH TIME ZONE CONSTRAINT users_updated_at_df DEFAULT CURRENT_TIMESTAMP
);