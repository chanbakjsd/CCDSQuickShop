CREATE TABLE admin_users (
	email TEXT UNIQUE NOT NULL
);
CREATE TABLE products (
	product_id         INTEGER PRIMARY KEY,
	name               TEXT NOT NULL,
	base_price         INTEGER NOT NULL,
	default_image_url  TEXT NOT NULL,
	-- JSON representing all the different variants.
	variants           TEXT NOT NULL,
	variant_image_urls TEXT NOT NULL,
	enabled            BOOLEAN NOT NULL
, sale_period
	INTEGER NOT NULL
	REFERENCES sale_periods(id)
	DEFAULT 1);
CREATE TABLE coupons (
	coupon_id             INTEGER PRIMARY KEY,
	coupon_code           TEXT NOT NULL,
	stripe_id             TEXT NOT NULL,
	min_purchase_quantity INTEGER,
	email_match           TEXT,
	discount_percentage   INTEGER NOT NULL,
	enabled               BOOLEAN NOT NULL,
	public                BOOLEAN NOT NULL,
	redemption_limit      INTEGER -- Not actually tracked by this server but Stripe instead.
, sale_period
	INTEGER NOT NULL
	REFERENCES sale_periods(id)
	DEFAULT 1);
CREATE TABLE orders (
	id                INTEGER PRIMARY KEY,
	order_id          TEXT UNIQUE NOT NULL,
	name              TEXT NOT NULL,
	matric_number     TEXT NOT NULL,
	email             TEXT NOT NULL,
	payment_reference TEXT UNIQUE,
	payment_time      DATETIME,
	collection_time   DATETIME,
	cancelled         BOOLEAN NOT NULL,
	coupon_id         INTEGER REFERENCES coupons(coupon_id)
, sale_period
	INTEGER NOT NULL
	REFERENCES sale_periods(id)
	DEFAULT 1);
CREATE TABLE order_items (
	order_id     TEXT    NOT NULL REFERENCES orders(order_id),
	product_id   TEXT    NOT NULL REFERENCES products(product_id),
	product_name TEXT    NOT NULL,
	unit_price   INTEGER NOT NULL,
	amount       INTEGER NOT NULL,
	image_url    TEXT    NOT NULL,
	-- JSON of the selected variants.
	variant      TEXT    NOT NULL
);
CREATE TABLE store_closures (
	id                INTEGER PRIMARY KEY,
	start_time        DATETIME NOT NULL,
	end_time          DATETIME NOT NULL,
	user_message      TEXT NOT NULL,
	allow_order_check BOOLEAN NOT NULL,
	deleted           BOOLEAN NOT NULL
);
CREATE TABLE IF NOT EXISTS "schema_migrations" (version varchar(128) primary key);
CREATE TABLE sale_periods (
	id          INTEGER PRIMARY KEY,
	admin_name  TEXT NOT NULL,
	start_time  DATETIME NOT NULL,
	delete_time DATETIME
);
-- Dbmate schema migrations
INSERT INTO "schema_migrations" (version) VALUES
  ('20250505031917'),
  ('20250505035817');
