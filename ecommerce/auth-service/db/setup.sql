DROP DATABASE IF EXISTS auth;
CREATE DATABASE auth;
USE auth;

DROP TABLE IF EXISTS roles;
CREATE TABLE roles (
    id INT PRIMARY KEY,
    name VARCHAR(20) NOT NULL UNIQUE
);

INSERT INTO roles (id, name) VALUES (0, 'UNSPECIFIED'), (1, 'CLIENT'), (2, 'ADMIN');

DROP TABLE IF EXISTS users;
CREATE TABLE users (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    username VARCHAR(32) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(64) NOT NULL UNIQUE,
    phone VARCHAR(20),
    role_id INT NOT NULL DEFAULT 1,  -- 1 = CLIENT
    FOREIGN KEY (role_id) REFERENCES roles(id)
) ENGINE=InnoDB;

SELECT 'Table users created successfully!' as status;