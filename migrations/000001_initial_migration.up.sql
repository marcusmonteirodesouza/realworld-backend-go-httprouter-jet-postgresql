CREATE TABLE IF NOT EXISTS users(
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     email TEXT NOT NULL,
     username TEXT NOT NULL,
     password_hash TEXT NOT NULL,
     bio TEXT,
     image TEXT,
     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
     updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);