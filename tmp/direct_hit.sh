#!/bin/bash

# Load .env variables
if [ -f "backend/.env" ]; then
    export $(grep -v '^#' backend/.env | xargs)
elif [ -f ".env" ]; then
    export $(grep -v '^#' .env | xargs)
fi

if [ -z "$META_ACCESS_TOKEN" ] || [ -z "$META_FACEBOOK_PAGE_ID" ]; then
    echo "Error: META_ACCESS_TOKEN or META_FACEBOOK_PAGE_ID not found."
    exit 1
fi

echo "--- PHASE 1: TESTING POST METRIC CANDIDATES ---"
# Get the latest post ID first
POST_ID=$(curl -s -G "https://graph.facebook.com/v22.0/${META_FACEBOOK_PAGE_ID}/posts" \
    -d "limit=1" \
    -d "access_token=${META_ACCESS_TOKEN}" | jq -r '.data[0].id')

if [ "$POST_ID" == "null" ] || [ -z "$POST_ID" ]; then
    echo "Error: Could not find any posts for Page ${META_FACEBOOK_PAGE_ID}"
    exit 1
fi

echo "Testing Post: $POST_ID"

echo -e "\nCandidate A: Classic Metrics (post_impressions_unique, post_engaged_users)"
curl -s -G "https://graph.facebook.com/v22.0/${POST_ID}/insights" \
    -d "metric=post_impressions_unique,post_engaged_users" \
    -d "access_token=${META_ACCESS_TOKEN}" | jq .

echo -e "\nCandidate B: Modern Metrics (post_reach, post_engagements)"
curl -s -G "https://graph.facebook.com/v22.0/${POST_ID}/insights" \
    -d "metric=post_reach,post_engagements" \
    -d "access_token=${META_ACCESS_TOKEN}" | jq .

echo -e "\nCandidate C: Simple Metadata (impressions, reach)"
curl -s -G "https://graph.facebook.com/v22.0/${POST_ID}/insights" \
    -d "metric=impressions,reach" \
    -d "access_token=${META_ACCESS_TOKEN}" | jq .

echo -e "\nCandidate D: Basic Engagement (post_reactions_by_type_total)"
curl -s -G "https://graph.facebook.com/v22.0/${POST_ID}/insights" \
    -d "metric=post_reactions_by_type_total" \
    -d "access_token=${META_ACCESS_TOKEN}" | jq .
