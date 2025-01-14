#!/usr/bin/env sh

export SESSION_SECRET
/app/ccds-shop \
	"-static=/app/static" \
	"-frontend=$FRONTEND_URL" \
	"-client-id=$GOOGLE_CLIENT_ID" \
	"-client-secret=$GOOGLE_CLIENT_SECRET" \
	"-stripe-secret=$STRIPE_SECRET_KEY" \
	"-stripe-webhook=$STRIPE_WEBHOOK_SECRET"
