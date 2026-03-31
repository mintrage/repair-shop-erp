CREATE TABLE parts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT,
    quantity BIGINT,
    price DECIMAL(19,4)
);

CREATE TABLE orders (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    client_name TEXT,
    phone_number TEXT,
    device_model TEXT,
    problem_description TEXT,
    status TEXT,
    creation_date TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);