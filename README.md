# VectorChat

A Go application that vectorizes text and files into PostgreSQL with pgvector, and uses these vectors to provide context for LLM interactions.

## Features

- Vectorize text and files using OpenAI embeddings
- Store vectors in PostgreSQL with pgvector extension
- Retrieve relevant context for LLM queries
- Simple HTTP API for chat interactions

## Prerequisites

- Docker and Docker Compose
- OpenAI API key

## Setup

1. Clone the repository:

   ```
   git clone https://github.com/devs-group/vectorchat.git
   cd vectorchat
   ```

2. Create a `.env` file with the required secrets:

   ```env
   OPENAI_API_KEY=your_openai_api_key
   GITHUB_CLIENT_ID=your_github_oauth_client_id
   GITHUB_CLIENT_SECRET=your_github_oauth_client_secret
   KRATOS_COOKIE_SECRET=at_least_32_random_bytes
   KRATOS_CIPHER_SECRET=at_least_32_random_bytes
   BASE_URL=localhost:4456
   FRONTEND_URL=localhost:3000
   LIGHT_FRONTEND_URL=localhost:3100
   VECTORCHAT_LIGHT_CLIENT_ID=your_vectorchat_light_client_id
   VECTORCHAT_LIGHT_CLIENT_SECRET=your_vectorchat_light_client_secret
   ```

   > See `docs/authentication.md` for a full breakdown of the authentication stack and additional environment variables.

3. Start the complete Docker stack:

   ```bash
   docker-compose up --build
   ```

4. Access the applications:

   - Main dashboard: http://localhost:3000
   - VectorChat Light: http://localhost:3100
   - API gateway (Oathkeeper): http://localhost:4456
   - Mailhog (dev email): http://localhost:8025

5. To interact with pgAdmin, open http://localhost:5050 and log in with the default email `admin@example.com` and password `admin`. The database password remains `postgres`.
