#!/usr/bin/env bash
# Linux / generic start (also used as Start.command sibling).
exec "$(cd "$(dirname "$0")" && pwd -P)/Start.command"
