
CREATE TABLE journals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL
        CHECK (length(btrim(title)) > 0),
    content TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_journals_user_created_at
    ON journals(user_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE TABLE journal_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_id UUID NOT NULL
        REFERENCES journals(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL
        CHECK (version_number > 0),
    title VARCHAR(200) NOT NULL
        CHECK (length(btrim(title)) > 0),
    content TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (journal_id, version_number)
);

CREATE INDEX idx_journal_versions_journal_created_at
    ON journal_versions(journal_id, created_at DESC);

CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(80) NOT NULL UNIQUE,
    slug VARCHAR(80) NOT NULL UNIQUE,
    is_system BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO categories (name, slug, is_system) VALUES
    ('Career / Work', 'career-work', TRUE),
    ('Learning', 'learning', TRUE),
    ('Finance', 'finance', TRUE),
    ('Daily Life', 'daily-life', TRUE);

CREATE TABLE journal_categories (
    journal_id UUID NOT NULL
        REFERENCES journals(id) ON DELETE CASCADE,
    category_id UUID NOT NULL
        REFERENCES categories(id) ON DELETE RESTRICT,
    source VARCHAR(10) NOT NULL DEFAULT 'manual'
        CHECK (source IN ('manual', 'ai')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (journal_id, category_id)
);

CREATE INDEX idx_journal_categories_category_id
    ON journal_categories(category_id);
