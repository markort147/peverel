\c peverel;

CREATE TABLE groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description VARCHAR(255),
    period INT NOT NULL,
    last_completed DATE,
    group_id INT,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE SET NULL
);

