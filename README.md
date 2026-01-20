# API Gateway

A high-performance, modular API Gateway built with Go, designed for microservices architectures. Features pluggable rate limiting strategies, API key management, and reverse proxy capabilities.

## Key Features

- **Flexible Rate Limiting**: Pluggable strategies (Token Bucket, Sliding Window, Fixed Window) with Redis-backed distributed storage
- **API Key Authentication**: Secure request validation with customizable rate limits per key
- **Reverse Proxy**: Intelligent request routing with load balancing support
- **Modular Architecture**: Interface-driven design for easy extension and testing
- **Production Ready**: Docker support, metrics collection, and comprehensive logging
- **Developer Friendly**: In-memory storage option for local development

## Use Cases

- Protect backend microservices from abuse and traffic spikes
- Manage API access with granular rate limiting per client
- Route requests across multiple service instances
- Monitor API usage and performance metrics

## Tech Stack

- Go 1.25.5
- Redis for distributed rate limiting
- PostgreSQL for API key storage
- Docker & Docker Compose
