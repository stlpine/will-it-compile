# will-it-compile Web Frontend

Modern React-based web interface for the **will-it-compile** service.

## 🚀 Quick Start

```bash
# Install dependencies
pnpm install

# Start development server (runs on port 3000)
pnpm dev

# Build for production
pnpm build

# Preview production build
pnpm preview

# Run tests
pnpm test
```

### Prerequisites

- **Node.js** 18+ with **Corepack** enabled
- **pnpm** 10+ (managed via Corepack)
- Backend API server running on `http://localhost:8080`

#### Setting up pnpm (first time)

This project uses **pnpm** via Corepack for faster installs and better disk efficiency:

```bash
# Enable Corepack (if not already enabled)
corepack enable

# The packageManager field in package.json will automatically use pnpm@10.22.0
```

### First Time Setup

1. **Install dependencies**:
   ```bash
   cd web
   pnpm install
   ```

2. **Configure environment** (optional):
   ```bash
   cp .env.example .env.local
   # Edit .env.local if needed
   ```

3. **Start backend API** (in another terminal):
   ```bash
   cd ..
   make run
   ```

4. **Start frontend dev server**:
   ```bash
   pnpm dev
   ```

5. **Open browser**: Navigate to `http://localhost:3000`

## 📁 Project Structure

```
web/
├── public/                  # Static assets
│   ├── index.html          # HTML template (not used - see index.html in root)
│   └── vite.svg            # Vite logo
│
├── src/
│   ├── components/         # Reusable UI components
│   │   ├── CodeEditor/     # ✅ Monaco-based code editor
│   │   ├── CompilerOutput/ # ✅ Compilation result display
│   │   ├── EnvironmentSelector/ # ✅ Language/compiler selector
│   │   └── JobStatus/      # ✅ Job status indicator
│   │
│   ├── pages/              # Page components
│   │   └── Home.tsx        # ✅ Main compilation page
│   │
│   ├── services/           # API clients
│   │   └── api.ts          # ✅ API client with polling & retry
│   │
│   ├── hooks/              # Custom React hooks
│   │   └── useCompilation.ts # ✅ Compilation logic hook
│   │
│   ├── utils/              # Utility functions
│   │   └── formatters.ts   # ✅ Output formatters
│   │
│   ├── styles/             # Global styles
│   │   └── globals.css     # ✅ Tailwind CSS with custom utilities
│   │
│   ├── types/              # TypeScript type definitions
│   │   └── api.ts          # ✅ API types (matches backend models)
│   │
│   ├── App.tsx             # ✅ Main app component
│   ├── main.tsx            # ✅ Entry point
│   └── vite-env.d.ts       # ✅ Vite type definitions
│
├── package.json            # ✅ Dependencies and scripts
├── tsconfig.json           # ✅ TypeScript configuration
├── vite.config.ts          # ✅ Vite config with API proxy
├── tailwind.config.js      # ✅ Tailwind CSS configuration
├── postcss.config.js       # ✅ PostCSS configuration
├── .eslintrc.cjs           # ✅ ESLint configuration
├── .prettierrc             # ✅ Prettier configuration
├── .env.example            # ✅ Environment variables template
└── README.md               # This file
```

## 🎯 Features

### ✅ Implemented (MVP)
- [x] Code editor with syntax highlighting (Monaco Editor)
- [x] Language/environment selector (C++, C, Go, Rust)
- [x] C++ standard selector (C++11, C++14, C++17, C++20, C++23)
- [x] Compile button with loading state
- [x] Compilation output display with formatting
- [x] Error and warning display
- [x] Real-time compilation status updates
- [x] Responsive design with Tailwind CSS
- [x] API client with error handling and retry logic
- [x] Job polling with exponential backoff
- [x] Base64 encoding for source code

### 🚧 Planned (Future Enhancements)
- [ ] Job history and management
- [ ] Multi-file project support (zip upload)
- [ ] Share compilation results (unique URLs)
- [ ] Dark/light theme toggle
- [ ] Integrated terminal
- [ ] Collaborative editing
- [ ] Template library (common code patterns)
- [ ] GitHub integration (compile from repo)
- [ ] Export results (PDF, markdown)

## 🛠️ Technology Stack

### Core
- **React 18** - UI framework
- **TypeScript** - Type safety
- **Vite** - Fast build tool and dev server

### UI Components
- **Monaco Editor** (`@monaco-editor/react`) - Code editor (same as VS Code)
- **Tailwind CSS** - Utility-first styling
- **Custom components** - Built from scratch with Tailwind

### HTTP & API
- **Axios** - HTTP client with interceptors
- **Custom polling** - Job status polling with exponential backoff

### Development Tools
- **ESLint** - Code linting
- **Prettier** - Code formatting
- **TypeScript** - Type checking

## 🔌 API Integration

The frontend communicates with the backend API server via Vite proxy.

### API Endpoints Used

```typescript
// Submit compilation job
POST /api/v1/compile
{
  "code": "base64_encoded_source",
  "language": "cpp",
  "compiler": "gcc-13",
  "standard": "c++20",
  "architecture": "x86_64",
  "os": "linux"
}

// Get job result
GET /api/v1/compile/:job_id

// List supported environments
GET /api/v1/environments

// Health check
GET /health
```

### Vite Proxy Configuration

The dev server proxies `/api` and `/health` to `http://localhost:8080`:

```typescript
// vite.config.ts
server: {
  port: 3000,
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true,
    },
    '/health': {
      target: 'http://localhost:8080',
      changeOrigin: true,
    },
  },
}
```

## 🎨 Design Guidelines

### UI Principles
1. **Simplicity**: Clean, uncluttered interface
2. **Responsiveness**: Works on desktop, tablet, mobile
3. **Performance**: Fast load times, smooth interactions
4. **Accessibility**: Semantic HTML, ARIA labels

### Component Patterns
- **Composition**: Small, reusable components
- **Props over State**: Minimize local state
- **Type Safety**: Strict TypeScript, no `any`
- **Tailwind CSS**: Utility-first styling

### Using shadcn/ui (Optional Enhancement)

This project uses Tailwind CSS for styling. To enhance the UI with shadcn/ui:

```bash
# Initialize shadcn/ui
npx shadcn-ui@latest init

# Add specific components
npx shadcn-ui@latest add button
npx shadcn-ui@latest add card
npx shadcn-ui@latest add select
npx shadcn-ui@latest add tabs
```

shadcn/ui provides beautifully designed, accessible components built with Radix UI and Tailwind CSS. See https://ui.shadcn.com/ for more information.

## 🔧 Development

### Environment Variables

Create `.env.local` file (optional):
```bash
# API Configuration
VITE_API_URL=/api/v1  # Default: uses Vite proxy
```

In production, set `VITE_API_URL` to your API server URL:
```bash
VITE_API_URL=https://api.example.com/api/v1
```

### Development Workflow

1. **Start backend API** (in terminal 1):
   ```bash
   cd /path/to/will-it-compile
   make run
   ```

2. **Start frontend dev server** (in terminal 2):
   ```bash
   cd web
   pnpm dev
   ```

3. **Access application**: `http://localhost:3000`

### Code Quality

```bash
# Lint code
pnpm lint

# Format code
pnpm format

# Type check
pnpm type-check
```

## 🧪 Testing

```bash
# Run unit tests
pnpm test

# Run with coverage
pnpm test:coverage
```

## 📦 Building for Production

```bash
# Create production build
pnpm build

# Preview production build
pnpm preview

# Build output location
ls -la dist/
```

The production build will be in the `dist/` directory.

### Deployment Options

1. **Static Hosting** (Vercel, Netlify, Cloudflare Pages)
   - Build command: `pnpm build`
   - Output directory: `dist`
   - Environment variables: Set `VITE_API_URL`

2. **Docker Container**
   ```dockerfile
   FROM node:18-alpine AS builder

   # Enable Corepack for pnpm
   RUN corepack enable

   WORKDIR /app

   # Copy package files
   COPY package.json pnpm-lock.yaml ./

   # Install dependencies
   RUN pnpm install --frozen-lockfile

   # Copy source and build
   COPY . .
   RUN pnpm build

   FROM nginx:alpine
   COPY --from=builder /app/dist /usr/share/nginx/html
   COPY nginx.conf /etc/nginx/conf.d/default.conf
   ```

3. **Kubernetes**
   - Use Helm chart in `/deployments/helm/`
   - Deploy alongside API server

## 🔐 Security Considerations

- **Input Sanitization**: Source code is Base64 encoded
- **CORS**: Configured in backend API
- **XSS Protection**: React escapes by default
- **No Secrets**: Environment variables for configuration only
- **Proxy in Dev**: Vite proxy prevents CORS issues locally

## 🐛 Common Issues

### Issue: pnpm install fails
**Solution**: Ensure you have Node.js 18+ and Corepack enabled (`corepack enable`). Try `pnpm store prune`

### Issue: API connection refused
**Solution**: Ensure backend API is running on `http://localhost:8080`

### Issue: CORS errors in production
**Solution**: Check API CORS configuration in `internal/api/middleware.go`

### Issue: Monaco Editor not loading
**Solution**: Vite handles Monaco web workers automatically. Ensure `@monaco-editor/react` is installed

### Issue: Compilation fails with "invalid base64"
**Solution**: Source code is automatically Base64 encoded in `Home.tsx:76`

## 📝 Implementation Notes

### Type Consistency

API types in `src/types/api.ts` match Go backend models in `pkg/models/`:
- `CompilationRequest` → `models.CompilationRequest`
- `CompilationResult` → `models.CompilationResult`
- `JobResponse` → `models.JobResponse`
- `Environment` → `models.Environment`

When updating backend models, update TypeScript types accordingly.

### Code Editor

The Monaco Editor component (`src/components/CodeEditor/CodeEditor.tsx`) maps language types:
- `cpp` / `c++` → Monaco `cpp`
- `c` → Monaco `c`
- `go` → Monaco `go`
- `rust` → Monaco `rust`

### Default Code Templates

Each language has a default "Hello, World!" template defined in `src/types/api.ts`:
```typescript
export const LANGUAGE_CONFIGS: Record<string, LanguageConfig> = {
  cpp: { ... defaultCode: "Hello World in C++" },
  c: { ... },
  go: { ... },
  rust: { ... },
}
```

## 📚 Resources

- [React Documentation](https://react.dev/)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [Tailwind CSS Docs](https://tailwindcss.com/docs)
- [Monaco Editor API](https://microsoft.github.io/monaco-editor/)
- [Vite Documentation](https://vitejs.dev/)
- [shadcn/ui Components](https://ui.shadcn.com/)
- [Backend API Documentation](../README.md)

## 🤝 Contributing

### Adding a New Language

1. **Update TypeScript types** (`src/types/api.ts`):
   ```typescript
   export type Language = 'c' | 'cpp' | 'go' | 'rust' | 'python'

   export const LANGUAGE_CONFIGS = {
     python: {
       language: 'python',
       label: 'Python',
       defaultCode: 'print("Hello, World!")',
       compiler: 'python3',
       fileExtension: 'py',
     },
   }
   ```

2. **Update CodeEditor mapping** (`src/components/CodeEditor/CodeEditor.tsx`):
   ```typescript
   function mapLanguage(language: Language): string {
     // ... existing cases
     case 'python':
       return 'python'
   }
   ```

3. **Test with backend**: Ensure backend supports the new language

### Adding a New Component

1. Create component directory:
   ```bash
   mkdir src/components/MyComponent
   ```

2. Create files:
   ```
   src/components/MyComponent/
   ├── MyComponent.tsx
   ├── index.ts
   ```

3. Use TypeScript and Tailwind CSS for styling

---

**Status**: ✅ **MVP Complete** - Fully implemented and ready to use
**Last Updated**: 2025-11-14
**Tech Stack**: React 18 + TypeScript + Vite + Tailwind CSS + Monaco Editor
**Maintainers**: See root README.md
