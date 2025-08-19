# dp-ecommerce-project
A microservices-based ecommerce platform built with Go, gRPC, and Docker

Project for the Master Course in Distributed Programming for Web, IoT and Mobile Systems  
@Unifi - MSc in Software: Science and Technology

---

# Prerequisite
- docker
- docker-compose
- make

---

# Run the project
*Quick start: clone the repository and start the services with Docker*
```bash
git clone https://github.com/marcoaga02/dp-ecommerce-project.git
cd dp-ecommerce-project
make build
make up
```

---

From folder's root (*dp-ecommerce-project*):
- `make build`
    - builds containers and compile protobuf files
- `make up` (or `make up_bg`, if you want to run the ecommerce in background)
    - to start all components
- `make stop`
    - to stop all docker container
- `make down`
    - to stop containers and to remove docker volumes and networks from the system

---

# The project
This repository contains an implementation of an ecommerce (online clothing store), in which clients can choose and buy clothes.  
There are two types of users:
- **Client**: a user who can 
    - browse the product catalogue;
    - add products to the shopping cart;
    - place oreders;
    - view the list and details of their orders;;
    - update quantities or remove products from the cart;
    - cancel an order (if not already canceled or delivered);
    - update their password and profile information.
- **Administrator**: a user with both admin and client privileges: 
    - *Admin functions*:
        - view the list of registered users;
        - change the role of other users (client <-> admin);
        - view the list of all orders;
        - see details of each order;
        - update order status (processing -> shipped -> delivered), but cannot cancel orders of other users.
    - *Client functions*: same as a regular client, so admins can also purchase products.

---

## Default Users and Products
When starting the ecommerce, default users and products are generated. Regarding users, two default users are created:
- an administrator:
    - username: admin
    - password: admin123
- a client:
    - username: demo-client
    - password: demo123

---

## How to use the ecommerce
- Register using the `Register` button in the top-right corner;
- Log in with one of the default accounts, or a new one you created, using the `Login` button;
- Enjoy shopping with the user-friendly interface.

## Some details about the architecture
This project is designed to consolidate knowledge of microservice architectures and web applications.  
Web clients communicate with the **Web Server**, which exposes HTTP routes bound to handler functions. Each handler interacts with the **Orchestrator**, which maps the HTTP request to the appropriate gRPC client.  
The gRPC clients obtain connections from the **Service Manager**, which periodically (every 15 seconds) closes and re-establishes connections to ensure reliability.  
Thanks to Docker Compose and gRPC health checks, if a server fails it is automatically restarted. The Service Manager then restores the connection, keeping the ecommerce system always available.  

![](ArchitecturalDiagram.png)

---

_Folder structure_
```plaintext
dp-ecommerce-project/  
|  
├── deploy/ # *.yml files for docker compose and .env file
|
├── ecommerce
|   ├── auth-service/ # module authentication microservice
|   ├── cart-service/ # module cart microservice
|   ├── order-service/ # module order microservice
|   ├── product-service/ # module product microservice
|   ├── orchestrator/ # module orchestrator (grpc clients, orchestrator, service manager, HTTP web server, ...)
|   ├── logger/ # module logger (custom logger implementation)
|   └── proto/ # *.proto files
|
├── Makefile
└── README.md
```
