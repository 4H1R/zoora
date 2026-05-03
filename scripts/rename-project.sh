#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "==> Replacing Go module path in all .go files..."
find "$ROOT" -name "*.go" -not -path "*/vendor/*" | xargs sed -i \
  's|github\.com/educonnect/backend|github.com/4H1R/zoora|g'

echo "==> Replacing in go.mod, go.sum..."
sed -i 's|github\.com/educonnect/backend|github.com/4H1R/zoora|g' \
  "$ROOT/go.mod" "$ROOT/go.sum" 2>/dev/null || true

echo "==> Replacing in swagger docs..."
for f in "$ROOT/docs/swagger.json" "$ROOT/docs/swagger.yaml"; do
  [ -f "$f" ] && sed -i \
    -e 's|github\.com/educonnect/backend|github.com/4H1R/zoora|g' \
    -e 's|github\.com/4H1R/educonnect-backend|github.com/4H1R/zoora|g' \
    "$f"
done

echo "==> Replacing TS type/interface names (GithubComEduconnectBackend -> GithubCom4H1RZoora)..."
find "$ROOT/frontend/src" -name "*.ts" -o -name "*.tsx" | xargs sed -i \
  -e 's|GithubComEduconnectBackend|GithubCom4H1RZoora|g' \
  -e 's|githubComEduconnectBackend|githubCom4H1RZoora|g'

echo "==> Renaming model files with old prefix..."
find "$ROOT/frontend/src/api/model" -name "githubComEduconnectBackend*.ts" | while read -r f; do
  dir="$(dirname "$f")"
  base="$(basename "$f")"
  newbase="${base/githubComEduconnectBackend/githubCom4H1RZoora}"
  mv "$f" "$dir/$newbase"
done

echo "==> Replacing EduConnect display name -> Zoora in i18n..."
for f in "$ROOT/frontend/src/i18n/locales/en.json" "$ROOT/frontend/src/i18n/locales/fa.json"; do
  [ -f "$f" ] && sed -i \
    -e 's|EduConnect|Zoora|g' \
    -e 's|Educonnect|Zoora|g' \
    -e 's|educonnect|zoora|g' \
    "$f"
done

echo "==> Replacing in docker-compose.yml (network names)..."
sed -i 's|educonnect|zoora|g' "$ROOT/docker-compose.yml"

echo "==> Replacing in .env.example..."
[ -f "$ROOT/.env.example" ] && sed -i \
  -e 's|EduConnect|Zoora|g' \
  -e 's|educonnect|zoora|g' \
  "$ROOT/.env.example"

echo "==> Replacing in CLAUDE.md files..."
for f in "$ROOT/CLAUDE.md" "$ROOT/frontend/CLAUDE.md" "$ROOT/.claude/skills/add-field/SKILL.md"; do
  [ -f "$f" ] && sed -i \
    -e 's|EduConnect|Zoora|g' \
    -e 's|educonnect-backend|zoora|g' \
    -e 's|educonnect|zoora|g' \
    "$f"
done

echo "==> Replacing git remote references in any config files..."
find "$ROOT" -name "*.json" -not -path "*/node_modules/*" -not -path "*/.git/*" | xargs grep -l "educonnect" 2>/dev/null | xargs sed -i \
  -e 's|4H1R/educonnect-backend|4H1R/zoora|g' \
  -e 's|educonnect/backend|4H1R/zoora|g' \
  -e 's|EduConnect|Zoora|g' \
  -e 's|GithubComEduconnectBackend|GithubCom4H1RZoora|g' \
  -e 's|githubComEduconnectBackend|githubCom4H1RZoora|g' \
  2>/dev/null || true

echo "==> Done. Verify with: grep -r 'educonnect\|EduConnect' --include='*.go' --include='*.ts' --include='*.tsx' ."
