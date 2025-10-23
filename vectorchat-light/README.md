# VectorChat Light

A lightweight Nuxt.js application that integrates with VectorChat API to create website assistant chatbots.

## Features

- Create chatbots trained on website content
- Automatic website crawling and indexing
- Clean, modern UI with Tailwind CSS
- TypeScript support

## Environment Configuration

Copy `.env.example` to `.env` and configure your VectorChat API credentials:

```bash
cp .env.example .env
```

Required environment variables:
- `NUXT_VECTORCHAT_URL`: VectorChat API base URL (e.g. `http://localhost:4456` when running through Oathkeeper)
- `NUXT_VECTORCHAT_CLIENT_ID`: OAuth2 client ID generated via VectorChat Swagger (`/auth/apikey`)
- `NUXT_VECTORCHAT_CLIENT_SECRET`: OAuth2 client secret paired with the client ID

## Setup

Make sure to install dependencies:

```bash
# npm
npm install

# pnpm
pnpm install

# yarn
yarn install

# bun
bun install
```

## Development Server

Start the development server on `http://localhost:3000`:

```bash
# npm
npm run dev

# pnpm
pnpm dev

# yarn
yarn dev

# bun
bun run dev
```

## Production

Build the application for production:

```bash
# npm
npm run build

# pnpm
pnpm build

# yarn
yarn build

# bun
bun run build
```

Locally preview production build:

```bash
# npm
npm run preview

# pnpm
pnpm preview

# yarn
yarn preview

# bun
bun run preview
```

Check out the [deployment documentation](https://nuxt.com/docs/getting-started/deployment) for more information.

## API Endpoints

### POST /api/chatbots

Creates a new chatbot and indexes a website for it.

**Request Body:**
```json
{
  "siteUrl": "https://example.com"
}
```

**Response:**
```json
{
  "chatbotId": "uuid-string",
  "siteUrl": "https://example.com",
  "previewUrl": "/preview/uuid-string",
  "message": "Chatbot created successfully with website content indexed."
}
```

## Integration with VectorChat

This application uses the VectorChat API to:
1. Create chatbots with optimized system prompts for website assistance
2. Upload and index website content for context-aware responses
3. Manage chatbot lifecycle with proper error handling and cleanup

For more information about VectorChat API, check the [documentation](https://nuxt.com/docs/getting-started/introduction).
