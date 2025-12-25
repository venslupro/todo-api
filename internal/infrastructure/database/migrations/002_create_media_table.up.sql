-- Create media_attachments table
CREATE TABLE media_attachments (
    id VARCHAR(36) PRIMARY KEY,
    todo_id VARCHAR(36) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_url TEXT NOT NULL,
    file_type INTEGER NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    thumbnail_url TEXT,
    duration INTEGER DEFAULT 0,
    uploaded_by VARCHAR(36) NOT NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraints
    CONSTRAINT fk_media_todo FOREIGN KEY (todo_id) REFERENCES todos(id) ON DELETE CASCADE,
    CONSTRAINT fk_media_user FOREIGN KEY (uploaded_by) REFERENCES users(id) ON DELETE CASCADE,
    
    -- Indexes for performance
    INDEX idx_media_todo_id (todo_id),
    INDEX idx_media_uploaded_by (uploaded_by),
    INDEX idx_media_uploaded_at (uploaded_at)
);

-- Create indexes for better query performance
CREATE INDEX idx_media_file_type ON media_attachments(file_type);
CREATE INDEX idx_media_created_at ON media_attachments(created_at);