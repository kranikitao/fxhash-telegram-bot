CREATE TABLE artist_subscriptions (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    fx_hash_artist_name text,
    fx_hash_artist_id text,
    chat_id bigint,
    is_active boolean
);

CREATE TABLE delivery_items (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    type text,
    chat_id bigint,
    generative_id bigint,
    generative_slug text,
    is_sent boolean,
    url text
);

CREATE TABLE events (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    chat_id bigint,
    event_code text,
    event_data text
);

CREATE TABLE subscribers (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    chat_id bigint,
    username text,
    subscribed boolean,
    state text,
    raw_user text
);

CREATE INDEX idx_artist_subscriptions_deleted_at ON artist_subscriptions USING btree (deleted_at);

CREATE INDEX idx_delivery_items_deleted_at ON delivery_items USING btree (deleted_at);

CREATE INDEX idx_events_deleted_at ON events USING btree (deleted_at);

CREATE INDEX idx_is_sent ON delivery_items USING btree (is_sent);

CREATE INDEX idx_subscribers_deleted_at ON subscribers USING btree (deleted_at);

CREATE UNIQUE INDEX uidx_chat_id ON subscribers USING btree (chat_id);

CREATE UNIQUE INDEX uidx_type_chat_id_generative_id ON delivery_items USING btree (type, chat_id, generative_id);