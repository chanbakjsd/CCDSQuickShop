-- migrate:up
CREATE TABLE sale_periods (
	id          INTEGER PRIMARY KEY,
	admin_name  TEXT NOT NULL,
	start_time  DATETIME NOT NULL,
	delete_time DATETIME
);

INSERT INTO sale_periods VALUES(1, 'Default', '2025-01-01 00:00:00', NULL);

ALTER TABLE products ADD COLUMN sale_period
	INTEGER NOT NULL
	REFERENCES sale_periods(id)
	DEFAULT 1;

ALTER TABLE coupons ADD COLUMN sale_period
	INTEGER NOT NULL
	REFERENCES sale_periods(id)
	DEFAULT 1;

ALTER TABLE orders ADD COLUMN sale_period
	INTEGER NOT NULL
	REFERENCES sale_periods(id)
	DEFAULT 1;

-- migrate:down
ALTER TABLE orders DROP COLUMN sale_period;
ALTER TABLE coupons DROP COLUMN sale_period;
ALTER TABLE products DROP COLUMN sale_period;
DROP TABLE sale_periods;
