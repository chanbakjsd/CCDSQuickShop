CREATE TABLE IF NOT EXISTS admin_users (
	email TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS products (
	product_id         INTEGER PRIMARY KEY,
	name               TEXT NOT NULL,
	base_price         INTEGER NOT NULL,
	default_image_url  TEXT NOT NULL,
	-- JSON representing all the different variants.
	variants           TEXT NOT NULL,
	variant_image_urls TEXT NOT NULL,
	enabled            BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS coupons (
	coupon_id             INTEGER PRIMARY KEY,
	coupon_code           TEXT NOT NULL,
	stripe_id             TEXT NOT NULL,
	min_purchase_quantity INTEGER,
	discount_percentage   INTEGER NOT NULL,
	enabled               BOOLEAN NOT NULL,
	public                BOOLEAN NOT NULL,
	redemption_limit      INTEGER -- Not actually tracked by this server but Stripe instead.
);

CREATE TABLE IF NOT EXISTS orders (
	id                INTEGER PRIMARY KEY,
	order_id          TEXT UNIQUE NOT NULL,
	name              TEXT NOT NULL,
	matric_number     TEXT NOT NULL,
	payment_reference TEXT UNIQUE,
	payment_time      DATETIME,
	collection_time   DATETIME,
	cancelled         BOOLEAN NOT NULL,
	coupon_id         INTEGER REFERENCES coupons(coupon_id)
);

CREATE TABLE IF NOT EXISTS order_items (
	order_id     TEXT    NOT NULL REFERENCES orders(order_id),
	product_id   TEXT    NOT NULL REFERENCES products(product_id),
	product_name TEXT    NOT NULL,
	unit_price   INTEGER NOT NULL,
	amount       INTEGER NOT NULL,
	image_url    TEXT    NOT NULL,
	-- JSON of the selected variants.
	variant      TEXT    NOT NULL
);
