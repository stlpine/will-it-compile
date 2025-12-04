# Documentation Index

Welcome to the **will-it-compile** documentation! This directory contains all project documentation organized by category.

## 📖 Quick Links

- [Main README](../README.md) - Project overview and getting started
- [Claude Code Context](../CLAUDE.md) - Context for AI-assisted development

## 📚 Documentation Structure

### 🏗️ Architecture

System design, architecture decisions, and deployment strategies:

- **[Implementation Plan](architecture/IMPLEMENTATION_PLAN.md)** - Detailed implementation roadmap and architecture
- **[Project Layout](architecture/PROJECT_LAYOUT.md)** - Project structure and organization patterns
- **[Kubernetes Architecture](architecture/KUBERNETES_ARCHITECTURE.md)** - Kubernetes deployment design and patterns

### 🔧 Development

Development guides, implementation details, and contribution guidelines:

- **[Testify Implementation](development/TESTIFY_IMPLEMENTATION.md)** - Testing framework and patterns

### 📘 User Guides

End-user documentation for CLI, TUI, and API:

- **[CLI Guide](guides/CLI_GUIDE.md)** - Command-line interface usage guide
- **[TUI Guide](guides/TUI_GUIDE.md)** - Terminal UI comprehensive guide

### 🔬 Technical Details

In-depth technical implementation documentation:

- **[Redis Integration](technical/REDIS_INTEGRATION.md)** - Redis storage setup and integration

## 🚀 Getting Started

### For Users
1. Read the [Main README](../README.md) for project overview
2. Check out the [CLI Guide](guides/CLI_GUIDE.md) or [TUI Guide](guides/TUI_GUIDE.md)

### For Developers
1. Review [Implementation Plan](architecture/IMPLEMENTATION_PLAN.md) for architecture
2. Check [Project Layout](architecture/PROJECT_LAYOUT.md) for code organization
3. Read [Testify Implementation](development/TESTIFY_IMPLEMENTATION.md) for testing patterns
4. See [CLAUDE.md](../CLAUDE.md) for AI-assisted development context

### For DevOps
1. Review [Kubernetes Architecture](architecture/KUBERNETES_ARCHITECTURE.md)
2. See [deployments/](../deployments/) for Helm charts and configs

## 📝 Documentation Conventions

- **Markdown format**: All documentation uses GitHub-flavored Markdown
- **Code examples**: Included where relevant with syntax highlighting
- **Version info**: Updated with significant changes
- **Cross-references**: Internal links use relative paths

## 🤝 Contributing to Documentation

When adding or updating documentation:

1. **Placement**: Choose the appropriate category directory
2. **Naming**: Use descriptive UPPERCASE names (e.g., `FEATURE_NAME_GUIDE.md`)
3. **Index**: Add your document to this index
4. **Links**: Update cross-references in related documents
5. **Format**: Follow existing markdown style and structure

## 📦 Additional Resources

- **Helm Charts**: See [deployments/helm/](../deployments/helm/)
- **Sample Code**: See [tests/samples/](../tests/samples/)
- **Scripts**: See [scripts/](../scripts/) for build and utility scripts

---

**Last Updated**: 2025-11-14
**Project Version**: MVP + Phase 2
