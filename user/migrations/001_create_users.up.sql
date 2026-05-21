CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
                       id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                       email         VARCHAR(255) NOT NULL UNIQUE,
                       password_hash VARCHAR(255) NOT NULL,
                       first_name    VARCHAR(100),
                       last_name     VARCHAR(100),
                       phone         VARCHAR(20),
                       role          VARCHAR(20) NOT NULL DEFAULT 'user',
                       created_at    TIMESTAMP NOT NULL DEFAULT now(),
                       updated_at    TIMESTAMP NOT NULL DEFAULT now()
);