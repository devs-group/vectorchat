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

2. Create a `.env` file with your OpenAI API key:

   ```
   OPENAI_API_KEY=your_openai_api_key
   ```

3. Build and start the application:

   ```
   make build
   make run
   ```

4. The application will be available at http://localhost:8080

5. To interact with pgadmin, access http://localhost:5050 and log in with the default password `admin` and the db password: `postgres`.
