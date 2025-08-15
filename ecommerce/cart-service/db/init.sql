DROP DATABASE IF EXISTS cart;
CREATE DATABASE cart;
USE cart;

DROP TABLE IF EXISTS cart_items;
CREATE TABLE cart_items (
    username VARCHAR(32) NOT NULL,
    code VARCHAR(32) NOT NULL,
    quantity INT UNSIGNED NOT NULL,
    PRIMARY KEY (username, code)
) ENGINE=InnoDB;