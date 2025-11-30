#!/bin/bash

# Migration helper script for Go-Auth project
# Usage: ./migrate.sh [up|down|status|help]

COMMAND=${1:-help}

case $COMMAND in
    up)
        echo "Running migrations..."
        go run ./cmd/main.go -migrate=up
        ;;
    down)
        echo "Rolling back last migration..."
        go run ./cmd/main.go -migrate=down
        ;;
    status)
        echo "Checking migration status..."
        go run ./cmd/main.go -migrate=status
        ;;
    help)
        echo "Migration Helper Script"
        echo ""
        echo "Usage: ./migrate.sh [command]"
        echo ""
        echo "Commands:"
        echo "  up      - Apply all pending migrations"
        echo "  down    - Rollback the last applied migration"
        echo "  status  - Show the status of applied migrations"
        echo "  help    - Show this help message"
        echo ""
        echo "Examples:"
        echo "  ./migrate.sh up"
        echo "  ./migrate.sh down"
        echo "  ./migrate.sh status"
        ;;
    *)
        echo "Unknown command: $COMMAND"
        echo "Use './migrate.sh help' for usage information"
        exit 1
        ;;
esac
