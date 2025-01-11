#!/usr/bin/env sh

SRC="$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))"

if [ -f "$SRC/.env" ]; then
	source "$SRC/.env"
fi

if [ -z "$GOOGLE_CLIENT_ID" ]; then
	echo -n "Google Client ID: "
	read GOOGLE_CLIENT_ID
fi
if [ -z "$GOOGLE_CLIENT_SECRET" ]; then
	echo -n "Google Client Secret: "
	read GOOGLE_CLIENT_SECRET
fi
if [ -z "$STRIPE_SECRET_KEY" ]; then
	echo -n "Stripe Secret Key: "
	read STRIPE_SECRET_KEY
fi
if [ -z "$STRIPE_WEBHOOK_SECRET" ]; then
	echo -n "Stripe Webhook Secret: "
	read STRIPE_WEBHOOK_SECRET
fi
if [ -z "$SESSION_SECRET" ]; then
	echo -n "Session Secret: "
	read SESSION_SECRET
fi

export SESSION_SECRET
go run . \
	-forward=http://localhost:5173 \
	"-client-id=$GOOGLE_CLIENT_ID" \
	"-client-secret=$GOOGLE_CLIENT_SECRET" \
	"-stripe-secret=$STRIPE_SECRET_KEY" \
	"-stripe-webhook=$STRIPE_WEBHOOK_SECRET"
