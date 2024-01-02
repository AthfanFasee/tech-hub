# My Blog Posts Platform (Back-End)

A technical hub platform composed of seven distinct services, each serving a unique purpose, and orchestrated through a broker service that acts as the sole gateway to the internet. The platform employs JWT tokens for user registration, login, and authentication, ensuring secure user interactions.

The platform features a payment service, enabling users to subscribe to a monthly plan, cancel subscriptions, and request refunds. A dedicated post service manages user posts, likes, dislikes, and comments, fostering real-time interactions. The platform also includes a logger service that logs crucial events to MongoDB, and a mail service that sends emails to users based on various events.

A listener service, powered by RabbitMQ, listens to events and reacts accordingly, contributing to the platform's event-driven architecture. The platform utilizes PostgreSQL and MySQL for storage, ensuring efficient data management. The entire project is powered by gRPC and Protocol Buffers, with gRPC metadata used as required.

The platform is Docker-powered, ensuring easy deployment and scalability. Exceptional error handling mechanisms are in place to ensure smooth user experience. The project has been optimized to eliminate performance bottlenecks, ensuring high performance and reliability. All services are designed with a focus on efficiency and robustness, making the platform technically strong and worth exploring.

## Technologies Used
Golang, Google Cloud, AWS, PostgreSQL, Docker, Postman

## Available Scripts
All the available commands are mentioned in Makefile in the root directory.
