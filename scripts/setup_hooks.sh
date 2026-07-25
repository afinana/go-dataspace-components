#!/usr/bin/env bash
set -e

HOOK_PATH=".git/hooks/pre-commit"

echo "Installing Git Pre-commit Quality Gate Hook..."

cat << 'EOF' > "$HOOK_PATH"
#!/usr/bin/env bash
set -e

echo "========================================================"
echo "🛡️  Running Dataspace Components Git Pre-Commit Quality Gate"
echo "========================================================"

make quality-gate

EOF

chmod +x "$HOOK_PATH"
echo "✅ Git pre-commit hook installed successfully at $HOOK_PATH!"
