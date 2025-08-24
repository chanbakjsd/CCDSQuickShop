# CCDS Quick Shop

CCDS Quick Shop is a quick implementation of merchandise store for
[NTU Students' Computing and Data Science Club](https://ntuscds.com). As of
writing this README, the live implementation can be found
[here](https://merch.ntuscds.com).

## Features

- Customizable Merch and Coupons without Restarting
- Simple Order Item Count Analytics
- Automatic Store Opening and Closure (sale period scheduling)
- Multiple Sale Periods ("seasons" of sales that are logically separate)
- Integration with Stripe for Payment Confirmation

## Screenshots

![](https://wenxu.dev/images/projects/CCDSQuickShop-User.png)
*Storefront Page for Users*

![](https://wenxu.dev/images/projects/CCDSQuickShop-Summary.png)
*Order Summary Page for Admin*

## Quick-Start

To run this program, you need [Go](https://go.dev) and
[pnpm](https://github.com/pnpm/pnpm).

The development mode can be run by using:

```sh
# Run both commands simultaneously.
$ pnpm dev

# In `backend/`:
$ ./dev.sh
```

A Dockerfile has also been provided that will create a Docker image that serves
both the built frontend and backend.

When starting the Docker image, configuration can be done using the following
environment variables:

- `FRONTEND_URL`: URL where the frontend can be accessed, used as base path
- `GOOGLE_CLIENT_ID`: Used for OAuth login of admin interface
- `GOOGLE_CLIENT_SECRET`: Used for OAuth login of admin interface
- `STRIPE_SECRET_KEY`: Secret key used to create orders and update coupons
  on Stripe
- `STRIPE_WEBHOOK_SECRET`: Secret of the Stripe webhook.

Stripe webhook is assumed to be configured to send requests to
`/api/v0/checkout/stripe`.
