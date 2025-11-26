CREATE TYPE goods AS (
    description VARCHAR(255) NOT NULL,
    price numeric(10,2) NOT NULL
);


CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    order INT NOT NULL UNIQUE,
    goods goods[] NOT NULL,
    status VARCHAR(10),
    accrual numeric(10,2)
);
