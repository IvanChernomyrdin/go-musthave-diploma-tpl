CREATE TYPE goods AS (
    description VARCHAR(255),
    price NUMERIC(10, 2)
);


CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL UNIQUE,
    goods goods[],
    status VARCHAR(10),
    accrual numeric(10,2)
);
