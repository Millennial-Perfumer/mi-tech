#!/bin/bash
# init-ssl.sh
# One-time SSL bootstrap script for mi-tech.millennialperfumer.in
# Run this manually via SSH on a fresh server (or when certs are missing).
# After this runs once, the certbot container handles all future renewals.

set -e

DOMAIN="mi-tech.millennialperfumer.in"
API_DOMAIN="mi-tech-api.millennialperfumer.in"
EMAIL="millennialperfumer.cc@gmail.com"   # ← Change this to your real email
CERT_PATH="./certbot/conf/live/$DOMAIN"
COMPOSE_FILE="docker-compose.prod.yml"

# Load .env file if it exists
if [ -f .env ]; then
  set -a
  source .env
  set +a
fi

if [ -z "$GITHUB_REPOSITORY_OWNER" ]; then
  echo "❌ Error: GITHUB_REPOSITORY_OWNER is not set."
  echo "   Please add it to your .env file or export it before running."
  exit 1
fi

if [ ! -f "$COMPOSE_FILE" ]; then
  echo "❌ Error: $COMPOSE_FILE not found."
  exit 1
fi

echo "==> [1/5] Creating certbot directory structure..."
mkdir -p ./certbot/conf/live/$DOMAIN
mkdir -p ./certbot/www

echo "==> [2/5] Generating temporary self-signed certificate so Nginx can start..."
openssl req -x509 -nodes -newkey rsa:2048 \
  -keyout $CERT_PATH/privkey.pem \
  -out $CERT_PATH/fullchain.pem \
  -days 1 \
  -subj "/CN=$DOMAIN"

# Create chain.pem (nginx config may reference it)
cp $CERT_PATH/fullchain.pem $CERT_PATH/chain.pem

echo "==> [3/5] Starting Nginx with temporary self-signed cert..."
docker compose -f $COMPOSE_FILE up -d nginx

echo "    Waiting for Nginx to be ready..."
sleep 10

echo "==> [4/5] Running Certbot to obtain real Let's Encrypt certificate..."
docker compose -f $COMPOSE_FILE run --rm certbot certonly \
  --webroot \
  --webroot-path=/var/www/certbot \
  --email $EMAIL \
  --agree-tos \
  --no-eff-email \
  -d $DOMAIN \
  -d $API_DOMAIN

echo "==> [5/5] Reloading Nginx with real certificate..."
docker compose -f $COMPOSE_FILE exec nginx nginx -s reload

echo ""
echo "✅ SSL bootstrap complete!"
echo "   https://$DOMAIN        → Frontend"
echo "   https://$API_DOMAIN    → API"
echo ""
echo "   The certbot container will now handle renewals every 12 hours automatically."
