#!/bin/bash

# Demo script for git-download-stats

set -e

PROG="./git-download-stats"

echo "================================"
echo "GitHub Download Stats Demo"
echo "================================"
echo ""

# Test 1: Fetch and display
echo "1️⃣  Fetch GitHub CLI stats (display only)"
echo "Command: $PROG fetch -o cli -r cli"
echo ""
$PROG fetch -o cli -r cli 2>&1 | head -15
echo "... (truncated)"
echo ""

# Test 2: Show latest from DB
echo "2️⃣  Show latest stored stats"
echo "Command: $PROG show cli cli"
echo ""
$PROG show cli cli 2>&1 | head -15
echo "... (truncated)"
echo ""

# Test 3: History
echo "3️⃣  Show statistics history"
echo "Command: $PROG history cli cli --limit 2"
echo ""
$PROG history cli cli --limit 2
echo ""

# Test 4: Database info
echo "4️⃣  Database information"
echo ""
echo "Database file: github-stats.db"
ls -lh github-stats.db
echo ""
echo "Records in database:"
sqlite3 github-stats.db "SELECT 'Total stats records:' as info, COUNT(*) as value FROM stats UNION ALL SELECT 'Total assets:', COUNT(*) FROM assets;"
echo ""

echo "================================"
echo "✅ Demo Complete!"
echo "================================"
echo ""
echo "Try these commands:"
echo "  $PROG fetch -o hashicorp -r terraform -s"
echo "  $PROG show hashicorp terraform"
echo "  $PROG history hashicorp terraform --limit 5"
echo "  $PROG compare hashicorp terraform --days 30"
echo ""
