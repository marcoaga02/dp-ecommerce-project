DROP DATABASE IF EXISTS shop_orders;
CREATE DATABASE shop_orders;
USE shop_orders;

DROP TABLE IF EXISTS statuses;
CREATE TABLE statuses (
    id INT PRIMARY KEY,
    name VARCHAR(20) NOT NULL UNIQUE
);

INSERT INTO statuses (id, name) VALUES
    (0, 'UNSPECIFIED'),
    (1, 'PROCESSING'),
    (2, 'SHIPPED'),
    (3, 'DELIVERED'),
    (4, 'CANCELED');

DROP TABLE IF EXISTS orders;
CREATE TABLE orders (
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(32) NOT NULL,
    status_id INT NOT NULL DEFAULT 1,  -- 1 = PROCESSING
    FOREIGN KEY (status_id) REFERENCES statuses(id)
) ENGINE=InnoDB;

DROP TABLE IF EXISTS order_items;
CREATE TABLE order_items (
    product_code VARCHAR(32) NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    unit_price DOUBLE(10,2) UNSIGNED NOT NULL,
    quantity INT UNSIGNED NOT NULL,
    order_id INT NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(id),
    PRIMARY KEY (product_code, order_id)
)