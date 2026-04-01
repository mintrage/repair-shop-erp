CREATE TABLE part (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT,
    quantity BIGINT,
    price DECIMAL(19,4)
);

CREATE TABLE repair_order (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    client_name TEXT,
    phone_number TEXT,
    device_model TEXT,
    problem_description TEXT,
    status TEXT,
    creation_date TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE repair_order_part (
    order_id BIGINT,
    part_id BIGINT,
    quantity BIGINT,
    price DECIMAL(19,4),
    PRIMARY KEY (order_id, part_id),
    CONSTRAINT fk_order FOREIGN KEY(order_id) REFERENCES repair_order(id),
    CONSTRAINT fk_part FOREIGN KEY(part_id) REFERENCES part(id)
);