# Contributing to ClaraVerse

First off, thank you for considering contributing to ClaraVerse! 🎉

ClaraVerse is built by the community, for the community. We welcome contributions from developers of all skill levels.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Commit Messages](#commit-messages)
- [Testing](#testing)

## Code of Conduct

This project adheres to a Code of Conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to [hello@claraverse.space](mailto:hello@claraverse.space).

**In Short:**
- Be respectful and inclusive
- Welcome newcomers and help them learn
- Accept constructive criticism gracefully
- Focus on what's best for the community
- Show empathy towards other community members

## How Can I Contribute?

### 🐛 Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates.

When creating a bug report, include:
- **Clear title and description**
- **Steps to reproduce** the issue
- **Expected vs actual behavior**
- **Screenshots** if applicable
- **Environment details**: OS, browser, versions
- **Error messages and logs**

### 💡 Suggesting Features

Feature suggestions are welcome! Please:
- **Check existing issues** for duplicates
- **Describe the problem** you're trying to solve
- **Explain your proposed solution** with examples
- **Consider alternatives** you've thought about
- **Explain why** this benefits the community

### 📝 Improving Documentation

Documentation improvements are always appreciated:
- Fix typos and grammar
- Add missing information
- Improve clarity and examples
- Add tutorials and guides
- Translate to other languages

### 🔧 Code Contributions

We especially welcome contributions in these areas:

- **Bug Fixes**: Check [open issues labeled "bug"](https://github.com/claraverse-space/ClaraVerse-Scarlet/labels/bug)
- **Features**: Check [open issues labeled "enhancement"](https://github.com/claraverse-space/ClaraVerse-Scarlet/labels/enhancement)
- **UI/UX**: Design improvements and accessibility
- **Testing**: Unit tests, integration tests, E2E tests
- **Integrations**: Connectors for new tools and services
- **Models**: Support for new LLM providers

## Development Setup

### Prerequisites

- **Go**: 1.24 or higher
- **Node.js**: 20 or higher
- **Python**: 3.11 or higher (for E2B service)
- **Docker**: Latest version (optional, for containerized development)
- **Git**: Latest version

### Quick Start

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/ClaraVerse-Scarlet.git
   cd ClaraVerse-Scarlet
   ```

2. **Install dependencies**
   ```bash
   make install
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   cp backend/providers.example.json backend/providers.json
   # Edit .env and providers.json with your configuration
   ```

4. **Start development environment**

   **Option A: Using tmux (recommended)**
   ```bash
   ./dev.sh
   ```

   **Option B: Manual (separate terminals)**
   ```bash
   # Terminal 1 - Backend
   cd backend
   go run ./cmd/server

   # Terminal 2 - Frontend
   cd frontend
   npm run dev

   # Terminal 3 - E2B Service (optional)
   cd backend/e2b-service
   python -m venv venv
   source venv/bin/activate
   pip install -r requirements.txt
   uvicorn main:app --reload --port 8001
   ```

5. **Access the application**
   - Frontend: http://localhost:5173
   - Backend: http://localhost:3001
   - E2B Service: http://localhost:8001

### Project Structure

```
ClaraVerse-Scarlet/
├── frontend/                 # React + TypeScript + Vite
│   ├── src/
│   │   ├── components/       # Reusable UI components
│   │   ├── pages/            # Page components
│   │   ├── services/         # API clients, WebSocket
│   │   ├── store/            # Zustand state management
│   │   └── utils/            # Utility functions
│   └── package.json
│
├── backend/                  # Go + Fiber API server
│   ├── cmd/server/           # Application entry point
│   ├── internal/
│   │   ├── handlers/         # HTTP & WebSocket handlers
│   │   ├── services/         # Business logic
│   │   ├── models/           # Data structures
│   │   └── middleware/       # Auth, CORS, etc.
│   ├── e2b-service/          # Python code execution service
│   └── docs/                 # Backend documentation
│
└── docs/                     # Project documentation
```

## Pull Request Process

### Before Submitting

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/bug-description
   ```

2. **Make your changes**
   - Write clean, maintainable code
   - Follow coding standards (see below)
   - Add tests for new functionality
   - Update documentation as needed

3. **Test your changes**
   ```bash
   # Frontend tests
   cd frontend
   npm run test
   npm run lint
   npm run type-check

   # Backend tests
   cd backend
   go test ./...
   go vet ./...
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add amazing feature"
   ```

### Submitting the PR

1. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Open a Pull Request**
   - Go to the [ClaraVerse repository](https://github.com/claraverse-space/ClaraVerse-Scarlet)
   - Click "New Pull Request"
   - Select your fork and branch
   - Fill out the PR template

3. **PR Title Format**
   ```
   type(scope): description

   Examples:
   feat(chat): add streaming message support
   fix(auth): resolve token refresh issue
   docs(readme): update installation instructions
   refactor(api): simplify error handling
   test(chat): add WebSocket connection tests
   ```

4. **PR Description Should Include**
   - **What** changed and **why**
   - **How** to test the changes
   - **Screenshots** for UI changes
   - **Breaking changes** if any
   - **Related issues** (e.g., "Fixes #123")

### Review Process

- Maintainers will review your PR within 2-3 business days
- Address feedback by pushing new commits
- Once approved, a maintainer will merge your PR
- Your contribution will be included in the next release!

## Coding Standards

### Frontend (TypeScript/React)

```typescript
// ✅ Good: Use functional components with TypeScript
interface ButtonProps {
  label: string;
  onClick: () => void;
  variant?: 'primary' | 'secondary';
}

export const Button: React.FC<ButtonProps> = ({ label, onClick, variant = 'primary' }) => {
  return (
    <button onClick={onClick} className={`btn btn-${variant}`}>
      {label}
    </button>
  );
};

// ❌ Bad: Avoid any types and unclear naming
const Btn = (props: any) => <button {...props} />;
```

**Guidelines:**
- Use TypeScript strict mode
- Prefer functional components with hooks
- Use meaningful variable names
- Extract reusable logic into custom hooks
- Keep components small and focused
- Use path aliases: `@/components` instead of `../../components`

### Backend (Go)

```go
// ✅ Good: Clear types, error handling, documentation
// HandleChatMessage processes incoming chat messages and streams responses
func (h *ChatHandler) HandleChatMessage(ctx *fiber.Ctx) error {
    var req ChatRequest
    if err := ctx.BodyParser(&req); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
    }

    // ... implementation
    return nil
}

// ❌ Bad: Unclear naming, missing error handling
func (h *ChatHandler) Msg(ctx *fiber.Ctx) error {
    var r ChatRequest
    ctx.BodyParser(&r)
    // ... implementation
}
```

**Guidelines:**
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use meaningful names (avoid abbreviations)
- Handle all errors explicitly
- Add comments for exported functions
- Use Go modules for dependencies
- Run `go fmt` and `go vet` before committing

### General Best Practices

- **DRY**: Don't Repeat Yourself - extract common logic
- **KISS**: Keep It Simple, Stupid - avoid over-engineering
- **YAGNI**: You Aren't Gonna Need It - don't add unused features
- **Security**: Validate inputs, sanitize outputs, use HTTPS
- **Performance**: Profile before optimizing, avoid premature optimization

## Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): subject

[optional body]

[optional footer]
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, no logic change)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks (dependencies, build, etc.)
- `perf`: Performance improvements
- `ci`: CI/CD changes

### Examples

```bash
feat(chat): add message editing functionality

fix(auth): resolve JWT token expiration issue
Fixes #456

docs(api): update WebSocket protocol documentation

refactor(store): simplify conversation state management

test(chat): add unit tests for message formatting

chore(deps): update React to v19
```

## Testing

### Frontend Tests

```bash
cd frontend

# Run all tests
npm run test

# Watch mode
npm run test:watch

# Coverage report
npm run test:coverage
```

**Testing Guidelines:**
- Write tests for new features
- Test edge cases and error conditions
- Use meaningful test descriptions
- Mock external dependencies
- Aim for >80% coverage on new code

### Backend Tests

```bash
cd backend

# Run all tests
go test ./...

# Verbose output
go test -v ./...

# Coverage report
go test -cover ./...
```

**Testing Guidelines:**
- Test business logic thoroughly
- Use table-driven tests for multiple scenarios
- Mock external services (LLM APIs, databases)
- Test error handling paths
- Include integration tests for critical flows

## Getting Help

- 💬 **Discord**: [Join our Discord](https://discord.gg/your-invite)
- 📧 **Email**: [hello@claraverse.space](mailto:hello@claraverse.space)
- 🐛 **Issues**: [GitHub Issues](https://github.com/claraverse-space/ClaraVerse-Scarlet/issues)
- 💡 **Discussions**: [GitHub Discussions](https://github.com/claraverse-space/ClaraVerse-Scarlet/discussions)

## Recognition

Contributors are recognized in:
- The [Contributors page](https://github.com/claraverse-space/ClaraVerse-Scarlet/graphs/contributors)
- Release notes for each version
- Our [website](https://claraverse.space) (for significant contributions)

## License

By contributing to ClaraVerse, you agree that your contributions will be licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**.

This ensures that:
- Your contributions remain free and open-source forever
- Companies using ClaraVerse must share their improvements
- The community benefits from all enhancements
- Developers always receive proper attribution

---

**Thank you for making ClaraVerse better! 🚀**

Every contribution, no matter how small, helps build a more private and powerful AI future.
