# will-it-compile Web Frontend

React-based web interface for the **will-it-compile** service.

## 🚀 Quick Start

```bash
# Install dependencies
npm install

# Start development server
npm start

# Build for production
npm run build

# Run tests
npm test
```

## 📁 Project Structure

```
web/
├── public/                  # Static assets
│   ├── index.html          # HTML template
│   └── favicon.ico         # Favicon
│
├── src/
│   ├── components/         # Reusable UI components
│   │   ├── CodeEditor/     # Code editor component
│   │   ├── CompilerOutput/ # Compilation result display
│   │   ├── EnvironmentSelector/ # Language/compiler selector
│   │   └── JobStatus/      # Job status indicator
│   │
│   ├── pages/              # Page components (routes)
│   │   ├── Home.tsx        # Main compilation page
│   │   ├── JobHistory.tsx  # Past compilations
│   │   └── About.tsx       # About page
│   │
│   ├── services/           # API clients and external services
│   │   ├── api.ts          # API client for backend
│   │   └── websocket.ts    # WebSocket for real-time updates
│   │
│   ├── hooks/              # Custom React hooks
│   │   ├── useCompilation.ts  # Compilation logic hook
│   │   └── useJobPolling.ts   # Job status polling hook
│   │
│   ├── utils/              # Utility functions
│   │   ├── formatters.ts   # Output formatters
│   │   └── validation.ts   # Input validation
│   │
│   ├── styles/             # Global styles and themes
│   │   ├── globals.css     # Global CSS
│   │   └── theme.ts        # Theme configuration
│   │
│   ├── types/              # TypeScript type definitions
│   │   └── api.ts          # API types (matches backend models)
│   │
│   ├── App.tsx             # Main app component
│   ├── index.tsx           # Entry point
│   └── setupTests.ts       # Test configuration
│
├── package.json            # Dependencies and scripts
├── tsconfig.json           # TypeScript configuration
└── README.md               # This file
```

## 🎯 Features (Planned)

### Phase 1: MVP
- [ ] Code editor with syntax highlighting
- [ ] Language/environment selector (C++, Go, Rust, Python)
- [ ] Compile button with loading state
- [ ] Compilation output display
- [ ] Error highlighting and parsing

### Phase 2: Enhanced UX
- [ ] Job history and management
- [ ] Real-time compilation status updates
- [ ] Multi-file project support (zip upload)
- [ ] Share compilation results (unique URLs)
- [ ] Dark/light theme toggle

### Phase 3: Advanced Features
- [ ] Integrated terminal
- [ ] Collaborative editing
- [ ] Template library (common code patterns)
- [ ] GitHub integration (compile from repo)
- [ ] Export results (PDF, markdown)

## 🛠️ Technology Stack

### Core
- **React 18** - UI framework
- **TypeScript** - Type safety
- **React Router** - Client-side routing

### UI Components
- **Monaco Editor** - Code editor (same as VS Code)
- **Tailwind CSS** - Utility-first styling
- **Radix UI** or **shadcn/ui** - Accessible component primitives

### State Management
- **Zustand** or **React Query** - State management and data fetching

### Testing
- **Vitest** - Fast unit test runner
- **React Testing Library** - Component testing
- **Playwright** - E2E testing

### Build Tools
- **Vite** - Fast build tool and dev server
- **ESLint** - Code linting
- **Prettier** - Code formatting

## 🔌 API Integration

The frontend communicates with the backend API server at `http://localhost:8080` (configurable).

### API Endpoints

```typescript
// Submit compilation job
POST /api/v1/compile
{
  "language": "cpp",
  "environment": "cpp-gcc-13",
  "source_code": "int main() { return 0; }"
}

// Get job result
GET /api/v1/compile/:job_id

// List supported environments
GET /api/v1/environments

// Health check
GET /health
```

### Example API Client

```typescript
// src/services/api.ts
import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

export const compileCode = async (request: CompilationRequest) => {
  const response = await axios.post(`${API_BASE_URL}/api/v1/compile`, request);
  return response.data;
};

export const getJobResult = async (jobId: string) => {
  const response = await axios.get(`${API_BASE_URL}/api/v1/compile/${jobId}`);
  return response.data;
};
```

## 🎨 Design Guidelines

### UI Principles
1. **Simplicity**: Clean, uncluttered interface
2. **Responsiveness**: Works on desktop, tablet, mobile
3. **Performance**: Fast load times, smooth interactions
4. **Accessibility**: WCAG 2.1 AA compliant

### Component Patterns
- **Atomic Design**: Atoms → Molecules → Organisms → Pages
- **Composition**: Small, reusable components
- **Props over State**: Minimize local state
- **Type Safety**: Strict TypeScript, no `any`

### Code Style
```typescript
// Good: Clear, typed, functional
interface CompileButtonProps {
  onCompile: (code: string) => Promise<void>;
  isLoading: boolean;
}

export const CompileButton: React.FC<CompileButtonProps> = ({
  onCompile,
  isLoading
}) => {
  return (
    <button
      onClick={() => onCompile(code)}
      disabled={isLoading}
      className="btn-primary"
    >
      {isLoading ? 'Compiling...' : 'Compile'}
    </button>
  );
};
```

## 🔧 Development

### Prerequisites
- Node.js 18+ and npm 9+
- Backend API server running on `http://localhost:8080`

### Environment Variables

Create `.env.local` file:
```bash
REACT_APP_API_URL=http://localhost:8080
REACT_APP_WS_URL=ws://localhost:8080
```

### Development Workflow

1. **Start backend API**:
   ```bash
   cd ..
   make run
   ```

2. **Start frontend dev server**:
   ```bash
   cd web
   npm start
   ```

3. **Access application**: `http://localhost:3000`

### Code Quality

```bash
# Lint code
npm run lint

# Format code
npm run format

# Type check
npm run type-check

# Run all checks
npm run validate
```

## 🧪 Testing

### Unit Tests
```bash
# Run all tests
npm test

# Run with coverage
npm run test:coverage

# Run specific test
npm test -- CodeEditor.test.tsx
```

### E2E Tests
```bash
# Run E2E tests
npm run test:e2e

# Run in headed mode (see browser)
npm run test:e2e:headed
```

### Test Structure
```
src/
├── components/
│   ├── CodeEditor/
│   │   ├── CodeEditor.tsx
│   │   └── CodeEditor.test.tsx
```

## 📦 Building for Production

```bash
# Create production build
npm run build

# Preview production build
npm run preview

# Build output location
ls -la build/
```

### Deployment Options

1. **Static Hosting** (Vercel, Netlify, Cloudflare Pages)
   - Connect GitHub repo
   - Auto-deploy on push

2. **Docker Container**
   - Dockerfile provided
   - Serves via nginx

3. **Kubernetes**
   - Helm chart in `/deployments/helm/`
   - Deployed alongside API

## 🔐 Security Considerations

- **Input Sanitization**: All user input is sanitized
- **CORS**: Configured for API domain only
- **CSP**: Content Security Policy headers
- **XSS Protection**: React escapes by default
- **No Secrets**: Never commit API keys or tokens

## 🤝 Contributing

### Adding a New Component

1. Create component directory:
   ```bash
   mkdir src/components/MyComponent
   ```

2. Create files:
   ```
   src/components/MyComponent/
   ├── MyComponent.tsx
   ├── MyComponent.test.tsx
   ├── MyComponent.module.css (if needed)
   └── index.ts (export)
   ```

3. Write tests first (TDD)
4. Implement component
5. Update Storybook (if using)

### Pull Request Checklist
- [ ] Tests pass (`npm test`)
- [ ] Linting passes (`npm run lint`)
- [ ] Type checking passes (`npm run type-check`)
- [ ] Components are documented
- [ ] Accessibility tested
- [ ] Responsive design verified

## 📚 Resources

- [React Documentation](https://react.dev/)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [Tailwind CSS Docs](https://tailwindcss.com/docs)
- [Monaco Editor API](https://microsoft.github.io/monaco-editor/)
- [Backend API Documentation](../docs/guides/API_GUIDE.md)

## 🐛 Common Issues

### Issue: API connection refused
**Solution**: Ensure backend API is running on `http://localhost:8080`

### Issue: CORS errors
**Solution**: Check API CORS configuration in `cmd/api/main.go`

### Issue: Monaco Editor not loading
**Solution**: Check webpack/vite configuration for web workers

## 📝 Notes

- This is a monorepo setup - frontend and backend share the same repository
- API types in `src/types/api.ts` should match `pkg/models/models.go`
- Use relative imports within web/, absolute for shared types
- Keep components small and focused (< 200 lines)

---

**Status**: 🚧 Planned (not yet implemented)
**Last Updated**: 2025-11-14
**Maintainers**: See root README.md
