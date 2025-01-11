-- name: CreateAdminUser :exec
INSERT INTO admin_users (
	email
) VALUES (
	?
) ON CONFLICT DO NOTHING;

-- name: CountAdminUsers :one
SELECT
	COUNT(*)
FROM
	admin_users;

-- name: ListAdminUsers :many
SELECT
	*
FROM
	admin_users;

-- name: AuthAdminUser :one
SELECT
	*
FROM
	admin_users
WHERE
	email = ?;

-- name: DeleteAdminUser :exec
DELETE FROM
	admin_users
WHERE
	email = ?;

-- name: CreateProduct :one
INSERT INTO products (
	name, base_price, default_image_url, variants, variant_image_urls, enabled
) VALUES (
	?, ?, ?, ?, ?, ?
) RETURNING product_id;

-- name: ListProducts :many
SELECT
	*
FROM
	products
WHERE
	enabled = TRUE
	OR CAST(@include_disabled AS BOOLEAN);

-- name: UpdateProduct :exec
UPDATE
	products
SET
	name = ?,
	base_price = ?,
	default_image_url = ?,
	variants = ?,
	variant_image_urls = ?,
	enabled = ?
WHERE
	product_id = ?;

-- name: SetProductEnabled :exec
UPDATE
	products
SET
	enabled = ?
WHERE
	product_id = ?;

-- name: CreateCoupon :one
INSERT INTO coupons (
	coupon_code, min_purchase_quantity, discount_percentage, enabled
) VALUES (
	?, ?, ?, ?
) RETURNING coupon_id;

-- name: ListPublicCoupons :many
SELECT
	*
FROM
	coupons
WHERE
	enabled = TRUE
	AND public = TRUE;

-- name: UseCoupon :one
SELECT
	*
FROM
	coupons
WHERE
	coupon_code = ?
	AND enabled = TRUE;

-- name: UpdateCoupon :exec
UPDATE
	coupons
SET
	coupon_code = ?,
	min_purchase_quantity = ?,
	discount_percentage = ?,
	enabled = ?,
	public = ?
WHERE
	coupon_id = ?;

-- name: SetCouponEnabled :exec
UPDATE
	coupons
SET
	enabled = ?
WHERE
	coupon_id = ?;

-- name: CreateOrder :one
INSERT INTO orders (
	name, matric_number, payment_reference, status, coupon_code
) VALUES (
	?, ?, ?, 'PLACED', ?
) RETURNING order_id;

-- name: LookupOrder :many
SELECT
	*
FROM
	orders
WHERE
	CAST(order_id AS TEXT) = ?
	OR matric_number = ? COLLATE NOCASE
	OR payment_reference = ? COLLATE NOCASE;

-- name: UpdateOrderStatusID :exec
UPDATE
	orders
SET
	status = ?
WHERE
	order_id = ?;

-- name: UpdateOrderStatusPaymentRef :exec
UPDATE
	orders
SET
	status = ?
WHERE
	payment_reference = ?;

-- name: CreateOrderItem :exec
INSERT INTO order_items (
	order_id, product_id, unit_price, amount, variant
) VALUES (
	?, ?, ?, ?, ?
);

-- name: ListOrderItems :many
SELECT
	*
FROM
	order_items
WHERE
	order_id = ?;