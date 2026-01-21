#!/bin/bash
# Lint-fix script for Unix/Linux/macOS - Tight feedback loop for fixing linting errors
# Run: ./scripts/unix/lint-fix.sh [path]
# Usage: Run this before committing to catch and fix linting errors quickly

set -e

WATCH=false
AUTOFIX=false
TARGET_PATH="."

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --watch|-w)
            WATCH=true
            shift
            ;;
        --autofix|-f)
            AUTOFIX=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS] [PATH]"
            echo ""
            echo "Options:"
            echo "  -w, --watch    Watch for file changes and re-lint"
            echo "  -f, --autofix  Auto-fix issues where possible"
            echo "  -h, --help     Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                    # Lint current directory"
            echo "  $0 ./internal         # Lint specific directory"
            echo "  $0 --watch            # Watch mode"
            echo "  $0 --autofix          # Auto-fix issues"
            exit 0
            ;;
        *)
            TARGET_PATH="$1"
            shift
            ;;
    esac
done

echo ""
echo "=== Lint-Fix: Tight Feedback Loop ==="
echo "Catch and fix linting errors before committing"
echo ""

# Get GOPATH
GOPATH=$(go env GOPATH)
if [ -z "$GOPATH" ]; then
    echo "Error: GOPATH not set"
    exit 1
fi

# Check if golangci-lint is installed
LINT_PATH="$GOPATH/bin/golangci-lint"

if [ ! -f "$LINT_PATH" ]; then
    echo "golangci-lint not found. Installing..."
    curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b "$GOPATH/bin" v2.8.0
    if [ $? -ne 0 ]; then
        echo "Failed to install golangci-lint."
        exit 1
    fi
fi

# Function to run linting
run_lint() {
    local autofix_flag=""
    if [ "$AUTOFIX" = true ]; then
        echo "[$(date +%H:%M:%S)] Attempting auto-fix for fixable issues..."
        autofix_flag="--fix"
    else
        echo "[$(date +%H:%M:%S)] Running lint check..."
    fi
    
    if $LINT_PATH run --timeout=10m --tests $autofix_flag "$TARGET_PATH" 2>&1 | tee /tmp/lint-output.txt; then
        echo ""
        echo "✅ Linting passed! Ready to commit."
        return 0
    else
        local exit_code=$?
        echo ""
        echo "❌ Linting failed with issues"
        echo ""
        echo "📋 Summary:"
        
        # Show top issues
        if grep -q "^\s\+\*" /tmp/lint-output.txt; then
            echo "Top issues:"
            grep "^\s\+\*" /tmp/lint-output.txt | head -10 | sed 's/^/  /'
        fi
        
        # Check for fixable issues
        if grep -qi "fixable\|can be auto-fixed" /tmp/lint-output.txt; then
            echo ""
            echo "💡 Some issues can be auto-fixed. Run with --autofix flag:"
            echo "   ./scripts/unix/lint-fix.sh --autofix"
        fi
        
        echo ""
        echo "💡 Tips:"
        echo "   - Run 'go fmt ./...' to fix formatting"
        echo "   - Run with --autofix to auto-fix what's possible"
        echo "   - Check full output above for details"
        
        return $exit_code
    fi
}

# Watch mode
if [ "$WATCH" = true ]; then
    echo "🔍 Watch mode: Monitoring for file changes..."
    echo "Press Ctrl+C to stop"
    echo ""
    
    # Check if fswatch or inotifywait is available
    if command -v fswatch &> /dev/null; then
        WATCH_CMD="fswatch"
    elif command -v inotifywait &> /dev/null; then
        WATCH_CMD="inotifywait"
    else
        echo "⚠️  Watch mode requires 'fswatch' (macOS) or 'inotifywait' (Linux)"
        echo "   Install: brew install fswatch (macOS) or apt-get install inotify-tools (Linux)"
        echo "   Falling back to single run..."
        WATCH=false
    fi
    
    if [ "$WATCH" = true ]; then
        # Run initial check
        run_lint
        
        # Watch for changes
        if [ "$WATCH_CMD" = "fswatch" ]; then
            fswatch -o "$TARGET_PATH" | while read; do
                sleep 0.5  # Debounce
                run_lint || true
            done
        elif [ "$WATCH_CMD" = "inotifywait" ]; then
            while inotifywait -r -e modify,create,delete "$TARGET_PATH" 2>/dev/null; do
                sleep 0.5  # Debounce
                run_lint || true
            done
        fi
    fi
fi

# Single run
if [ "$WATCH" = false ]; then
    if ! run_lint; then
        echo ""
        echo "⚠️  Fix errors before committing, or use 'git commit --no-verify' to skip"
        exit 1
    fi
fi
