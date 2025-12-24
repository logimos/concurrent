#!/usr/bin/env bash
set -euo pipefail

# init_docs.sh
# Sets up MkDocs + GitHub Actions deploy to gh-pages for a repo.
# - Auto-detects repo info from git when possible
# - Handles mkdocs.yml at root or in subdirectory
# - Preserves existing requirements.txt dependencies
# - Creates/updates:
#   - mkdocs.yml (if missing)
#   - requirements.txt (merges with existing if present)
#   - .github/workflows/docs.yml
#   - docs/index.md (if missing)
#
# NOTE: This script does NOT configure GitHub repo settings for you (Pages settings must be done in the UI).

say() { printf "%s\n" "$*"; }
err() { printf "ERROR: %s\n" "$*" >&2; }
warn() { printf "WARNING: %s\n" "$*" >&2; }

prompt() {
  local var="$1" msg="$2" def="${3:-}"
  local input=""
  if [[ -n "$def" ]]; then
    read -r -p "${msg} [${def}]: " input
    input="${input:-$def}"
  else
    read -r -p "${msg}: " input
  fi
  printf -v "$var" "%s" "$input"
}

yesno() {
  local var="$1" msg="$2" def="${3:-y}"
  local input=""
  while true; do
    read -r -p "${msg} [y/n] (default: ${def}): " input
    input="${input:-$def}"
    case "${input,,}" in
      y|yes) printf -v "$var" "y"; return 0;;
      n|no)  printf -v "$var" "n"; return 0;;
      *) say "Please answer y or n.";;
    esac
  done
}

# -------- Auto-detect from git --------
detect_git_info() {
  local remote_url=""
  local default_branch=""
  
  if command -v git >/dev/null 2>&1; then
    # Detect default branch
    if default_branch=$(git symbolic-ref --short HEAD 2>/dev/null); then
      : # Got current branch
    elif default_branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null); then
      : # Got current branch
    else
      default_branch="main"
    fi
    
    # Detect repo owner/name from remote
    if remote_url=$(git remote get-url origin 2>/dev/null); then
      # Handle both https:// and git@ formats
      if [[ "$remote_url" =~ github\.com[:/]([^/]+)/([^/]+)(\.git)?$ ]]; then
        DETECTED_OWNER="${BASH_REMATCH[1]}"
        DETECTED_REPO="${BASH_REMATCH[2]%.git}"
      fi
    fi
  fi
  
  DETECTED_BRANCH="${default_branch:-main}"
}

# -------- Detect mkdocs.yml location --------
detect_mkdocs_location() {
  if [[ -f "mkdocs.yml" ]]; then
    MKDOCS_AT_ROOT=true
    MKDOCS_DIR="."
  elif [[ -f "mkdocs.yaml" ]]; then
    MKDOCS_AT_ROOT=true
    MKDOCS_DIR="."
    warn "Found mkdocs.yaml (not .yml) - script expects .yml"
  else
    MKDOCS_AT_ROOT=false
    MKDOCS_DIR=""
  fi
}

# -------- Initialize --------
say ""
say "MkDocs -> GitHub Pages (gh-pages branch) setup"
say "=============================================="
say ""

# Check if we're in a git repo
if [[ ! -d ".git" ]]; then
  err "This doesn't look like a git repo (no .git directory). Run from the repo root."
  exit 1
fi

# Auto-detect git info
DETECTED_OWNER=""
DETECTED_REPO=""
DETECTED_BRANCH="main"
detect_git_info

# Detect mkdocs.yml location
MKDOCS_AT_ROOT=false
MKDOCS_DIR=""
detect_mkdocs_location

# -------- Prompts --------
if [[ "$MKDOCS_AT_ROOT" == "true" ]]; then
  say "✓ Detected mkdocs.yml at repository root"
  DOCS_ROOT="."
else
  prompt DOCS_ROOT "Docs root folder (relative to repo root, or '.' for root)" "."
fi

prompt ACTIONS_DIR "Location of GitHub Actions workflows dir" ".github/workflows"

# Determine if material theme is already used
USE_MATERIAL="n"
if [[ -f "${DOCS_ROOT}/mkdocs.yml" ]] && grep -q "name: material" "${DOCS_ROOT}/mkdocs.yml" 2>/dev/null; then
  USE_MATERIAL="y"
  say "✓ Detected mkdocs-material theme in existing mkdocs.yml"
else
  yesno USE_MATERIAL "Add mkdocs-material theme?" "y"
fi

# Repo info prompts with defaults from git
if [[ -n "$DETECTED_OWNER" ]] && [[ -n "$DETECTED_REPO" ]]; then
  prompt REPO_NAME "Repo name (as on GitHub)" "$DETECTED_REPO"
  prompt OWNER_NAME "GitHub username or org (owner)" "$DETECTED_OWNER"
else
  prompt REPO_NAME "Repo name (as on GitHub, e.g. my-repo)"
  prompt OWNER_NAME "GitHub username or org (owner), e.g. username"
fi

# Normalize: strip trailing slashes, handle "." as root
DOCS_ROOT="${DOCS_ROOT%/}"
if [[ "$DOCS_ROOT" == "." ]] || [[ -z "$DOCS_ROOT" ]]; then
  DOCS_ROOT="."
  MKDOCS_DIR="."
else
  MKDOCS_DIR="$DOCS_ROOT"
fi
ACTIONS_DIR="${ACTIONS_DIR%/}"

# -------- Derived paths --------
if [[ "$DOCS_ROOT" == "." ]]; then
  MKDOCS_YML="mkdocs.yml"
  REQ_TXT="requirements.txt"
  WORKING_DIR="."
else
  MKDOCS_YML="${DOCS_ROOT}/mkdocs.yml"
  REQ_TXT="${DOCS_ROOT}/requirements.txt"
  WORKING_DIR="${DOCS_ROOT}"
fi

DOCS_MD_DIR="${DOCS_ROOT}/docs"
INDEX_MD="${DOCS_MD_DIR}/index.md"
WORKFLOW_FILE="${ACTIONS_DIR}/docs.yml"

SITE_URL="https://${OWNER_NAME}.github.io/${REPO_NAME}/"

# -------- Create directories --------
mkdir -p "${DOCS_MD_DIR}"
mkdir -p "${ACTIONS_DIR}"

# -------- Create mkdocs.yml if missing --------
if [[ -f "${MKDOCS_YML}" ]]; then
  say "✓ Found existing ${MKDOCS_YML} - leaving it as-is."
else
  say "Creating ${MKDOCS_YML}"
  if [[ "${USE_MATERIAL}" == "y" ]]; then
    cat > "${MKDOCS_YML}" <<EOF
site_name: ${REPO_NAME} Docs
site_url: ${SITE_URL}

theme:
  name: material

nav:
  - Home: index.md

markdown_extensions:
  - admonition
  - toc:
      permalink: true
EOF
  else
    cat > "${MKDOCS_YML}" <<EOF
site_name: ${REPO_NAME} Docs
site_url: ${SITE_URL}

theme:
  name: mkdocs

nav:
  - Home: index.md
EOF
  fi
fi

# -------- Handle requirements.txt intelligently --------
say "Updating ${REQ_TXT}"

# Required dependencies
REQUIRED_DEPS=()
if [[ "${USE_MATERIAL}" == "y" ]]; then
  REQUIRED_DEPS=("mkdocs>=1.5" "mkdocs-material>=9.5")
else
  REQUIRED_DEPS=("mkdocs>=1.5")
fi

# Read existing requirements if present
EXISTING_DEPS=()
if [[ -f "${REQ_TXT}" ]]; then
  say "  Found existing ${REQ_TXT} - merging dependencies"
  while IFS= read -r line || [[ -n "$line" ]]; do
    # Skip empty lines and comments
    [[ -z "$line" ]] && continue
    [[ "$line" =~ ^[[:space:]]*# ]] && continue
    # Extract package name (before >=, ==, etc.)
    if [[ "$line" =~ ^([^>=!<[:space:]]+) ]]; then
      EXISTING_DEPS+=("$line")
    fi
  done < "${REQ_TXT}"
fi

# Merge dependencies: keep existing, add required if missing
declare -A DEPS_MAP
# Add existing deps
for dep in "${EXISTING_DEPS[@]}"; do
  pkg_name="${dep%%[>=!<]*}"
  pkg_name="${pkg_name// /}"
  DEPS_MAP["$pkg_name"]="$dep"
done

# Add/update required deps
for dep in "${REQUIRED_DEPS[@]}"; do
  pkg_name="${dep%%[>=!<]*}"
  DEPS_MAP["$pkg_name"]="$dep"
done

# Write merged requirements.txt
{
  echo "# MkDocs dependencies"
  echo "# Generated by init_docs.sh"
  echo ""
  for dep in "${DEPS_MAP[@]}"; do
    echo "$dep"
  done | sort
} > "${REQ_TXT}"

say "  ✓ Updated ${REQ_TXT} with merged dependencies"

# -------- Create docs index if missing --------
if [[ -f "${INDEX_MD}" ]]; then
  say "✓ Found existing ${INDEX_MD} - leaving it as-is."
else
  say "Creating ${INDEX_MD}"
  cat > "${INDEX_MD}" <<EOF
# ${REPO_NAME} documentation

Welcome 👋

- This site is built with **MkDocs**.
- It is deployed automatically to **GitHub Pages** whenever files in \`${DOCS_ROOT}\` change.

## Next steps

- Edit this page: \`${INDEX_MD}\`
- Add more pages under: \`${DOCS_MD_DIR}/\`
- Update navigation in: \`${MKDOCS_YML}\`
EOF
fi

# -------- Create GitHub Actions workflow --------
say "Creating/updating ${WORKFLOW_FILE}"

# Determine paths for workflow triggers
if [[ "$DOCS_ROOT" == "." ]]; then
  # Build paths section inline
  PATHS_SECTION="      - \"mkdocs.yml\"
      - \"mkdocs.yaml\"
      - \"requirements.txt\"
      - \"docs/**\"
      - \"${ACTIONS_DIR}/docs.yml\""
else
  PATHS_SECTION="      - \"${DOCS_ROOT}/**\"
      - \"${ACTIONS_DIR}/docs.yml\""
fi

cat > "${WORKFLOW_FILE}" <<EOF
name: Deploy docs

on:
  push:
    branches: [ "${DETECTED_BRANCH}" ]
    paths:
${PATHS_SECTION}

permissions:
  contents: write

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Configure Git credentials
        run: |
          git config user.name github-actions[bot]
          git config user.email 41898282+github-actions[bot]@users.noreply.github.com

      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: "3.x"

      - name: Install MkDocs dependencies
        run: |
          python -m pip install --upgrade pip
          pip install -r "${REQ_TXT}"

      - name: Deploy to GitHub Pages (gh-pages branch)
        working-directory: "${WORKING_DIR}"
        run: python -m mkdocs gh-deploy --force
EOF

# -------- Helpful output --------
say ""
say "✅ Setup complete! Files created/updated:"
say "  - ${MKDOCS_YML}"
say "  - ${REQ_TXT}"
say "  - ${INDEX_MD}"
say "  - ${WORKFLOW_FILE}"
say ""
say "Local test (optional, recommended):"
if [[ "$WORKING_DIR" != "." ]]; then
  say "  cd ${WORKING_DIR}"
fi
say "  python3 -m pip install -r ${REQ_TXT}"
say "  python3 -m mkdocs serve"
say ""
say "Commit & push:"
if [[ "$DOCS_ROOT" == "." ]]; then
  say "  git add mkdocs.yml ${REQ_TXT} ${WORKFLOW_FILE}"
else
  say "  git add ${DOCS_ROOT} ${WORKFLOW_FILE}"
fi
say "  git commit -m \"Add MkDocs + GitHub Pages deploy\""
say "  git push"
say ""
say "Then in GitHub UI (required):"
say "  1) Repo -> Settings -> Pages"
say "  2) Source: Deploy from a branch"
say "  3) Branch: gh-pages"
say "  4) Folder: / (root)"
say ""
say "Your docs URL will be:"
say "  ${SITE_URL}"
say ""
say "Trigger behavior:"
say "  - Only pushes that change relevant files redeploy the docs."
say "  - Default branch detected: ${DETECTED_BRANCH}"
say ""
say "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
say ""
say "📋 Next Steps for ${REPO_NAME} Repository:"
say ""
say "  1️⃣  Test locally (recommended):"
say "     └─ cd ${WORKING_DIR}"
say "     └─ python3 -m pip install -r ${REQ_TXT}"
say "     └─ python3 -m mkdocs serve"
say "     └─ Open http://127.0.0.1:8000 in your browser"
say ""
say "  2️⃣  Commit and push the changes:"
say "     └─ git add ${MKDOCS_YML} ${REQ_TXT} ${WORKFLOW_FILE}"
if [[ -f "${INDEX_MD}" ]] && ! grep -q "Welcome 👋" "${INDEX_MD}" 2>/dev/null; then
  say "     └─ git add ${INDEX_MD}"
fi
say "     └─ git commit -m \"docs: add MkDocs GitHub Pages deployment\""
say "     └─ git push origin ${DETECTED_BRANCH}"
say ""
say "  3️⃣  Configure GitHub Pages (one-time setup):"
say "     └─ Go to: https://github.com/${OWNER_NAME}/${REPO_NAME}/settings/pages"
say "     └─ Source: Deploy from a branch"
say "     └─ Branch: gh-pages"
say "     └─ Folder: / (root)"
say "     └─ Click Save"
say ""
say "  4️⃣  Verify deployment:"
say "     └─ Wait for GitHub Actions workflow to complete"
say "     └─ Check Actions tab: https://github.com/${OWNER_NAME}/${REPO_NAME}/actions"
say "     └─ Visit your docs: ${SITE_URL}"
say ""
say "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
say ""
say "💡 Tips:"
say "  • The workflow will automatically deploy when you push changes to:"
say "    - ${MKDOCS_YML}"
say "    - ${REQ_TXT}"
say "    - Files in ${DOCS_MD_DIR}/"
say "  • Edit your docs in ${DOCS_MD_DIR}/ and update navigation in ${MKDOCS_YML}"
say "  • The gh-pages branch will be created automatically on first deployment"
say ""

