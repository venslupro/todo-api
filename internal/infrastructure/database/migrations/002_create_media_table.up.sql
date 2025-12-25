-- Create media table
CREATE TABLE media
(
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    todo_id       UUID NOT NULL,
    file_name     VARCHAR(255) NOT NULL,
    file_url      TEXT         NOT NULL,
    file_type     INTEGER      NOT NULL,
    file_size     BIGINT       NOT NULL,
    mime_type     VARCHAR(100) NOT NULL,
    thumbnail_url TEXT,
    duration      INTEGER                  DEFAULT 0,
    uploaded_by   UUID NOT NULL,
    uploaded_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- Foreign key constraints
    CONSTRAINT fk_media_todo FOREIGN KEY (todo_id) REFERENCES todos (id) ON DELETE CASCADE,
    CONSTRAINT fk_media_user FOREIGN KEY (uploaded_by) REFERENCES users (id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX idx_media_todo_id ON media (todo_id);
CREATE INDEX idx_media_uploaded_by ON media (uploaded_by);
CREATE INDEX idx_media_uploaded_at ON media (uploaded_at);
CREATE INDEX idx_media_file_type ON media (file_type);
CREATE INDEX idx_media_created_at ON media (created_at);