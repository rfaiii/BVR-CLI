-- Schema for the local product catalog used by the LOOKUP IMAGE feature.
-- This is a standalone SQLite database (default path: ./data/catalog.db).

PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS products (
    id                  TEXT PRIMARY KEY,
    brand               TEXT NOT NULL,
    line                TEXT,
    model               TEXT,
    style_code          TEXT,
    category            TEXT NOT NULL,
    subcategory         TEXT,
    title               TEXT NOT NULL,
    description         TEXT,
    material            TEXT,
    color               TEXT,
    pattern             TEXT,
    hardware            TEXT,
    dimensions_text     TEXT,
    width_cm            REAL,
    height_cm           REAL,
    depth_cm            REAL,
    handle_drop_cm      REAL,
    closure_type        TEXT,
    interior_features   TEXT,
    exterior_features   TEXT,
    era                 TEXT,
    gender              TEXT,
    reference_url       TEXT,
    source_name         TEXT,
    source_license      TEXT,
    notes               TEXT,
    created_at          TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS product_aliases (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id  TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    alias       TEXT NOT NULL,
    UNIQUE(product_id, alias)
);

CREATE TABLE IF NOT EXISTS product_images (
    id          TEXT PRIMARY KEY,
    product_id  TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    file_path   TEXT,
    image_url   TEXT,
    view_type   TEXT,
    sha256      TEXT,
    width       INTEGER,
    height      INTEGER,
    caption     TEXT
);

CREATE TABLE IF NOT EXISTS product_embeddings (
    product_id      TEXT PRIMARY KEY REFERENCES products(id) ON DELETE CASCADE,
    model           TEXT NOT NULL,
    embedding_json  TEXT NOT NULL,
    searchable_text TEXT NOT NULL,
    updated_at      TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS lookup_feedback (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    lookup_id           TEXT,
    selected_product_id TEXT,
    correct_product_id  TEXT,
    feedback_type       TEXT NOT NULL,
    notes               TEXT,
    created_at          TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS image_lookup_cache (
    image_sha256       TEXT NOT NULL,
    vision_model       TEXT NOT NULL,
    prompt_version     TEXT NOT NULL,
    analysis_json      TEXT NOT NULL,
    created_at         TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (image_sha256, vision_model, prompt_version)
);

CREATE VIRTUAL TABLE IF NOT EXISTS products_fts USING fts5(
    product_id UNINDEXED,
    brand,
    line,
    model,
    style_code,
    category,
    material,
    color,
    pattern,
    searchable_text
);
