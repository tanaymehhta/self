# Self Frontend

Next.js 14 web application for the Self digital memory assistant.

## Features

- **Modern UI**: shadcn/ui components with Tailwind CSS
- **Real-time Updates**: WebSocket integration for live transcription
- **Audio Streaming**: WebRTC for real-time audio processing
- **Responsive Design**: Works on desktop, tablet, and mobile
- **Authentication**: Secure JWT-based authentication
- **Dark Mode**: System preference and manual toggle

## Tech Stack

- **Framework**: Next.js 14 with App Router
- **Styling**: Tailwind CSS + shadcn/ui components
- **State Management**: Zustand + React Query
- **Animation**: Framer Motion
- **Audio**: Web Audio API + WebRTC
- **Real-time**: Socket.io client
- **Forms**: React Hook Form + Zod validation

## Getting Started

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Start production server
npm start
```

## Environment Variables

Create `.env.local`:

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
NEXT_PUBLIC_MINIO_URL=http://localhost:9000
```

## Project Structure

```
frontend/
├── app/                    # Next.js app directory
│   ├── (auth)/            # Auth-protected routes
│   ├── (public)/          # Public routes
│   └── api/               # API routes
├── components/            # React components
│   ├── ui/                # shadcn/ui components
│   ├── audio/             # Audio-related components
│   ├── chat/              # Chat interface
│   └── dashboard/         # Dashboard components
├── hooks/                 # Custom React hooks
├── lib/                   # Utilities and configurations
├── store/                 # Zustand stores
└── types/                 # TypeScript type definitions
```

## Key Components

### Audio Recording
- Real-time audio capture
- Waveform visualization
- Recording controls

### Chat Interface
- Natural language queries
- Conversation history
- Real-time responses

### Timeline View
- Chronological conversation display
- File interaction overlay
- Interactive navigation

### Dashboard
- Activity overview
- Quick insights
- Recent conversations

## Development

### Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run start` - Start production server
- `npm run lint` - Run ESLint
- `npm run type-check` - Run TypeScript checks

### Testing

- `npm test` - Run Jest tests
- `npm run test:watch` - Run tests in watch mode
- `npm run test:coverage` - Generate coverage report

### Code Quality

- ESLint for linting
- Prettier for formatting
- TypeScript for type safety
- Husky for git hooks