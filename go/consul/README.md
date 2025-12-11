# Consul

To learn Consul.

- User Service (Port 8081): Manages user data, communicates with Order Service
- Order Service (Port 8082): Manages orders, publishes events to RabbitMQ
- PostgreSQL: Separate databases for each service (user_db on 5432, order_db on 5433)
- RabbitMQ: Message queue for async communication (Port 5672, Management UI on 15672)
- Consul: Service discovery and health checks (Port 8500)
