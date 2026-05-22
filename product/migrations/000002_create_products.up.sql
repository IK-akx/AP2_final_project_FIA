CREATE TABLE products (
                          id UUID PRIMARY KEY,
                          name VARCHAR(255) NOT NULL,
                          description TEXT,

                          category_id UUID REFERENCES categories(id) ON DELETE SET NULL,

                          price DECIMAL(10,2) NOT NULL CHECK (price >= 0),

                          stock INT NOT NULL DEFAULT 0 CHECK (stock >= 0),

                          manufacturer VARCHAR(255),

                          requires_rx BOOLEAN DEFAULT false,

                          image_url VARCHAR(500),

                          created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                          updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);