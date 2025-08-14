DROP DATABASE IF EXISTS product;
CREATE DATABASE product;
USE product;

DROP TABLE IF EXISTS sizes;
CREATE TABLE sizes (
    id INT PRIMARY KEY,
    name VARCHAR(20) NOT NULL UNIQUE
);

INSERT INTO sizes (id, name) VALUES
    (0, 'UNSPECIFIED'),
    (1, 'XS'),
    (2, 'S'),
    (3, 'M'),
    (4, 'L'),
    (5, 'XL'),
    (6, 'XXL');

DROP TABLE IF EXISTS products;
CREATE TABLE products (
    code VARCHAR(32) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    size_id INT NOT NULL,
    color VARCHAR(30) NOT NULL,
    description VARCHAR(255) NOT NULL,
    stock INT NOT NULL,
    price DOUBLE(10,2) NOT NULL,
    FOREIGN KEY (size_id) REFERENCES sizes(id)
) ENGINE=InnoDB;