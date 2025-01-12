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

-- name: CouponByID :one
SELECT
	*
FROM
	coupons
WHERE
	coupon_id = ?;

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

-- name: CreateOrder :exec
INSERT INTO orders (
	order_id, name, matric_number, email, payment_reference, payment_time, collection_time, cancelled, coupon_id
) VALUES (
	?, ?, ?, ?, NULL, NULL, NULL, FALSE, ?
);

-- name: AssociateOrder :exec
UPDATE
	orders
SET
	payment_reference = ?
WHERE
	order_id = ?;

-- name: LookupOrder :many
SELECT
	*
FROM
	orders
WHERE
	(
		order_id = @id COLLATE NOCASE
		OR matric_number = @id COLLATE NOCASE
		OR payment_reference = @id COLLATE NOCASE
		OR email = @id COLLATE NOCASE
		OR email = @id || '@e.ntu.edu.sg' COLLATE NOCASE
	)
	AND
	(
		CAST(@include_cancelled AS BOOLEAN)
		OR order_id = @id COLLATE NOCASE -- Always allow lookup via order ID even if it's cancelled.
		OR cancelled = FALSE
	);

-- name: CompleteCheckout :one
UPDATE
	orders
SET
	payment_time = COALESCE(payment_time, ?),
	coupon_id = (
		SELECT * FROM (
			SELECT
				coupon_id
			FROM
				coupons
			WHERE
				stripe_id = @coupon_stripe_id
			UNION ALL
			SELECT NULL
		) x
		LIMIT 1
	)
WHERE
	orders.payment_reference = ?
RETURNING order_id;

-- name: ExpireCheckout :one
UPDATE
	orders
SET
	cancelled = TRUE
WHERE
	payment_reference = ?
RETURNING order_id;


-- name: UpdateCollectionTime :exec
UPDATE
	orders
SET
	collection_time = COALESCE(collection_time, ?)
WHERE
	order_id = ?;

-- name: UpdateCancelled :one
UPDATE
	orders
SET
	cancelled = ?
WHERE
	order_id = ?
	AND cancelled = FALSE
RETURNING
	payment_reference;

-- name: CreateOrderItem :exec
INSERT INTO order_items (
	order_id, product_id, product_name, unit_price, amount, image_url, variant
) VALUES (
	?, ?, ?, ?, ?, ?, ?
);

-- name: ListOrderItems :many
SELECT
	*
FROM
	order_items
WHERE
	order_id = ?;
