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
	stripe_id, coupon_code, min_purchase_quantity, email_match, discount_percentage, enabled, public
) VALUES (
	?, ?, ?, ?, ?, ?, ?
) RETURNING coupon_id;

-- name: ListCoupons :many
SELECT
	*
FROM
	coupons;

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

-- name: CouponEnabledByCode :one
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
	stripe_id = ?,
	coupon_code = ?,
	min_purchase_quantity = ?,
	email_match = ?,
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

-- name: LookupOrderFromItem :many
SELECT
	orders.*
FROM
	orders
WHERE
	EXISTS(
		SELECT
			1
		FROM
			order_items
		WHERE
			order_items.order_id = orders.order_id
			AND order_items.product_name = @product_name COLLATE NOCASE
			AND order_items.variant = @variant COLLATE NOCASE
	)
	AND orders.collection_time IS NULL
	AND orders.payment_time IS NOT NULL
	AND orders.cancelled = FALSE;

-- name: UnfulfilledOrderSummary :many
SELECT
	order_items.product_id, order_items.product_name, order_items.variant,
	SUM(order_items.amount)
FROM
	orders
	JOIN order_items ON orders.order_id = order_items.order_id
WHERE
	orders.collection_time IS NULL
	AND orders.payment_time IS NOT NULL
	AND orders.cancelled = FALSE
GROUP BY
	order_items.product_id, order_items.product_name, order_items.variant;

-- name: UnfulfilledOrderIDs :many
SELECT
	order_id
FROM
	orders
WHERE
	orders.collection_time IS NULL
	AND orders.payment_time IS NOT NULL
	AND orders.cancelled = FALSE
LIMIT
	@max_count;

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

-- name: CreateStoreClosure :one
INSERT INTO store_closures (
	start_time, end_time, user_message, allow_order_check, deleted
) VALUES (
	?, ?, ?, ?, FALSE
) RETURNING id;

-- name: StoreClosureCurrent :one
SELECT
	*
FROM
	store_closures
WHERE
	@current_time >= start_time
	AND @current_time <= end_time
	AND deleted = FALSE
LIMIT 1;

-- name: ListStoreClosures :many
SELECT
	*
FROM
	store_closures
WHERE
	deleted = FALSE;

-- name: UpdateStoreClosure :exec
UPDATE
	store_closures
SET
	start_time = ?,
	end_time = ?,
	user_message = ?,
	allow_order_check = ?
WHERE
	id = ?
	AND deleted = FALSE;

-- name: DeleteStoreClosure :exec
UPDATE
	store_closures
SET
	deleted = TRUE
WHERE
	id = ?;
