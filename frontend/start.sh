#!/bin/sh
set -eu

: "${API_BASE:?API_BASE env var must be set}"

HTML_FILE=/usr/share/nginx/html/index.html

if [ -f "$HTML_FILE" ]; then
  sed -i "s|__API_BASE__|${API_BASE}|g" "$HTML_FILE"
  echo "40-inject-env.sh: injected API_BASE=${API_BASE} into ${HTML_FILE}"
else
  echo "40-inject-env.sh: warning, ${HTML_FILE} not found" >&2
fi
