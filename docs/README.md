# Documentation

This directory contains the MkDocs documentation for the `concurrent` package.

## Building Documentation

### Prerequisites

Install Python dependencies:

```bash
pip install -r requirements.txt
```

### Development Server

Start the MkDocs development server:

```bash
mkdocs serve
```

The documentation will be available at `http://127.0.0.1:8000`

### Build Static Site

Build the documentation as a static site:

```bash
mkdocs build
```

The built site will be in the `site/` directory.

### Deploy

Deploy to GitHub Pages:

```bash
mkdocs gh-deploy
```

## Documentation Structure

- `index.md` - Homepage
- `getting-started/` - Installation and quick start guides
- `features/` - Detailed feature documentation
- `examples/` - Code examples
- `api/` - API reference

