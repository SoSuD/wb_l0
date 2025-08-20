CREATE TABLE orders (
                        order_uid        TEXT        PRIMARY KEY,
                        track_number     TEXT        NOT NULL,
                        entry            TEXT        NOT NULL,
                        locale           TEXT,
                        internal_signature TEXT,
                        customer_id      TEXT,
                        delivery_service TEXT,
                        shardkey         TEXT,
                        sm_id            INTEGER,
                        date_created     TIMESTAMPTZ NOT NULL,
                        oof_shard        TEXT
);

CREATE INDEX idx_orders_track_number ON orders (track_number);
CREATE INDEX idx_orders_customer_id  ON orders (customer_id);
CREATE INDEX idx_orders_date_created ON orders (date_created);


CREATE TABLE deliveries (
                            order_uid     TEXT  PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
                            name      TEXT NOT NULL,
                            phone     TEXT,
                            zip       TEXT,
                            city      TEXT NOT NULL,
                            address   TEXT NOT NULL,
                            region    TEXT,
                            email     TEXT
);


CREATE TABLE payments (
                          order_uid     TEXT  PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
                          transaction    TEXT  NOT NULL UNIQUE,
                          request_id     TEXT,
                          currency       TEXT  NOT NULL,
                          provider       TEXT,
                          amount         INTEGER NOT NULL CHECK (amount >= 0),
                          payment_dt     BIGINT NOT NULL,
                          bank           TEXT,
                          delivery_cost  INTEGER NOT NULL CHECK (delivery_cost >= 0),
                          goods_total    INTEGER NOT NULL CHECK (goods_total  >= 0),
                          custom_fee     INTEGER NOT NULL CHECK (custom_fee    >= 0)
);


CREATE TABLE items (
                       order_uid     TEXT   NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
                       chrt_id       BIGINT NOT NULL,
                       track_number  TEXT   NOT NULL,
                       price         INTEGER NOT NULL CHECK (price >= 0),
                       rid           TEXT   NOT NULL,
                       name          TEXT   NOT NULL,
                       sale          INTEGER NOT NULL CHECK (sale >= 0),
                       size          TEXT   NOT NULL,
                       total_price   INTEGER NOT NULL CHECK (total_price >= 0),
                       nm_id         BIGINT NOT NULL,
                       brand         TEXT   NOT NULL,
                       status        INTEGER NOT NULL,
                       PRIMARY KEY (order_uid, chrt_id)
);

CREATE INDEX idx_items_order_id ON items(order_uid);
