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
   git clone https://github.com/yourusername/vectorchat.git
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

## Usage

### Upload Files

Upload a file to be vectorized and used for chat context:

```bash
curl -X POST http://localhost:8080/upload \
  -F "file=@/path/to/your/file.txt" \
  -F "chat_id=my-chat-session"
```

This will return a chat_id that you can use for subsequent chat requests.

### Manage Files

#### List files in a chat session:

```bash
curl http://localhost:8080/files/my-chat-session
```

#### Update a file:

```bash
curl -X PUT http://localhost:8080/files/my-chat-session/file.txt \
  -F "file=@/path/to/updated-file.txt"
```

#### Delete a file:

```bash
curl -X DELETE http://localhost:8080/files/my-chat-session/file.txt
```

### Chat with Context

Send a chat message and get a response with context from your uploaded files:

```bash
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"chat_id":"my-chat-session","query":"What is the main topic of the document?"}'
```

Or use the convenience command:

```bash
make test-api
```

### Adding Documents

Documents are automatically added to the vector database when the application starts. To add more documents, you can modify the `addDemoDocument` function in `cmd/main.go`.

## Testing

### Run tests locally:

```bash
make test
```

### Run tests inside the Docker container:

```bash
make test-in-container
```

### Run integration tests:

```bash
make test-integration-in-container
```

### Run database tests:

```bash
make test-db-in-container
```

## Cleanup

Stop the application:

```bash
make stop
```

Clean up everything including volumes:

```bash
make clean
```

## Project Structure

- `cmd/main.go`: Application entry point
- `pkg/db`: Database interaction with pgvector
- `pkg/vectorize`: Text and file vectorization
- `pkg/chat`: Chat service with context from vector database
- `integration_test.go`: Integration tests

## License

MIT
