CREATE TABLE IF NOT EXISTS chats (
    id BIGINT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS links (
    id BIGSERIAL PRIMARY KEY,
    url TEXT NOT NULL UNIQUE,
    last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS chat_links (
    chat_id BIGINT NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    link_id BIGINT NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    PRIMARY KEY (chat_id, link_id)
);

CREATE TABLE IF NOT EXISTS tags (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS link_tags (
    chat_id BIGINT NOT NULL,
    link_id BIGINT NOT NULL,
    tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (chat_id, link_id, tag_id),
    FOREIGN KEY (chat_id, link_id) REFERENCES chat_links(chat_id, link_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_chat_links_link_id ON chat_links(link_id);
