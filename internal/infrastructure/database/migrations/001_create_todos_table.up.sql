-- Enable UUID extension
CREATE
EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable pg_trgm extension for full-text search
CREATE
EXTENSION IF NOT EXISTS pg_trgm;

-- Users table
CREATE TABLE IF NOT EXISTS users
(
    id
    UUID
    PRIMARY
    KEY
    DEFAULT
    gen_random_uuid
(
),
    email VARCHAR
(
    255
) UNIQUE NOT NULL,
    username VARCHAR
(
    100
) UNIQUE NOT NULL,
    password_hash VARCHAR
(
    255
) NOT NULL,
    full_name VARCHAR
(
    255
),
    avatar_url TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP
                         WITH TIME ZONE DEFAULT NOW()
    );

-- TODO table (updated structure)
CREATE TABLE IF NOT EXISTS todos
(
    id
    UUID
    PRIMARY
    KEY
    DEFAULT
    gen_random_uuid
(
),
    user_id UUID NOT NULL REFERENCES users
(
    id
) ON DELETE CASCADE,
    title VARCHAR
(
    500
) NOT NULL,
    description TEXT,
    status INTEGER NOT NULL DEFAULT 1, -- Maps to common.v1.Status enum
    priority INTEGER NOT NULL DEFAULT 2, -- Maps to common.v1.Priority enum
    due_date TIMESTAMP
  WITH TIME ZONE,
      tags TEXT[] DEFAULT '{}',
      is_shared BOOLEAN DEFAULT FALSE,
      shared_by UUID REFERENCES users(id),
    created_at TIMESTAMP
  WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP
  WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP
  WITH TIME ZONE,
      assigned_to UUID REFERENCES users(id),
    parent_id UUID REFERENCES todos
(
    id
)
  ON DELETE CASCADE,
    position INTEGER NOT NULL DEFAULT 0
    );

-- Teams table
CREATE TABLE IF NOT EXISTS teams
(
    id
    UUID
    PRIMARY
    KEY
    DEFAULT
    gen_random_uuid
(
),
    name VARCHAR
(
    255
) NOT NULL,
    description TEXT,
    created_by UUID NOT NULL REFERENCES users
(
    id
),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP
                         WITH TIME ZONE DEFAULT NOW()
    );

-- Team members table
CREATE TABLE IF NOT EXISTS team_members
(
    team_id
    UUID
    NOT
    NULL
    REFERENCES
    teams
(
    id
) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users
(
    id
)
  ON DELETE CASCADE,
    role INTEGER NOT NULL DEFAULT 1, -- Maps to common.v1.Role enum
    joined_at TIMESTAMP
  WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY
(
    team_id,
    user_id
)
    );

-- Shared TODOs table
CREATE TABLE IF NOT EXISTS shared_todos
(
    id
    UUID
    PRIMARY
    KEY
    DEFAULT
    gen_random_uuid
(
),
    todo_id UUID NOT NULL REFERENCES todos
(
    id
) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams
(
    id
)
  ON DELETE CASCADE,
    shared_by UUID NOT NULL REFERENCES users
(
    id
),
    shared_at TIMESTAMP
  WITH TIME ZONE DEFAULT NOW(),
    UNIQUE
(
    todo_id,
    team_id
)
    );

-- Media files table
CREATE TABLE IF NOT EXISTS media
(
    id
    UUID
    PRIMARY
    KEY
    DEFAULT
    gen_random_uuid
(
),
    todo_id UUID NOT NULL REFERENCES todos
(
    id
) ON DELETE CASCADE,
    file_name VARCHAR
(
    500
) NOT NULL,
    file_url TEXT NOT NULL,
    file_type INTEGER NOT NULL, -- Maps to common.v1.MediaType enum
    file_size BIGINT NOT NULL,
    mime_type VARCHAR
(
    100
),
    thumbnail_url TEXT,
    duration INTEGER, -- For videos
    uploaded_by UUID NOT NULL REFERENCES users
(
    id
),
    uploaded_at TIMESTAMP
  WITH TIME ZONE DEFAULT NOW()
    );

-- Activity logs table
CREATE TABLE IF NOT EXISTS activity_logs
(
    id
    UUID
    PRIMARY
    KEY
    DEFAULT
    gen_random_uuid
(
),
    team_id UUID REFERENCES teams
(
    id
) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users
(
    id
),
    action VARCHAR
(
    50
) NOT NULL,
    resource_type VARCHAR
(
    50
) NOT NULL,
    resource_id UUID,
    details JSONB,
    created_at TIMESTAMP
  WITH TIME ZONE DEFAULT NOW()
    );

-- Indexes for users table
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- Indexes for todos table
CREATE INDEX IF NOT EXISTS idx_todos_user_id ON todos(user_id);
CREATE INDEX IF NOT EXISTS idx_todos_status ON todos(status);
CREATE INDEX IF NOT EXISTS idx_todos_priority ON todos(priority);
CREATE INDEX IF NOT EXISTS idx_todos_due_date ON todos(due_date);
CREATE INDEX IF NOT EXISTS idx_todos_created_at ON todos(created_at);
CREATE INDEX IF NOT EXISTS idx_todos_assigned_to ON todos(assigned_to);
CREATE INDEX IF NOT EXISTS idx_todos_parent_id ON todos(parent_id);
CREATE INDEX IF NOT EXISTS idx_todos_tags ON todos USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_todos_title_trgm ON todos USING gin(title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_todos_description_trgm ON todos USING gin(description gin_trgm_ops);

-- Indexes for teams table
CREATE INDEX IF NOT EXISTS idx_teams_created_by ON teams(created_by);

-- Indexes for team_members table
CREATE INDEX IF NOT EXISTS idx_team_members_team_id ON team_members(team_id);
CREATE INDEX IF NOT EXISTS idx_team_members_user_id ON team_members(user_id);

-- Indexes for shared_todos table
CREATE INDEX IF NOT EXISTS idx_shared_todos_todo_id ON shared_todos(todo_id);
CREATE INDEX IF NOT EXISTS idx_shared_todos_team_id ON shared_todos(team_id);

-- Indexes for media table
CREATE INDEX IF NOT EXISTS idx_media_todo_id ON media(todo_id);
CREATE INDEX IF NOT EXISTS idx_media_uploaded_by ON media(uploaded_by);

-- Indexes for activity_logs table
CREATE INDEX IF NOT EXISTS idx_activity_logs_team_id ON activity_logs(team_id);
CREATE INDEX IF NOT EXISTS idx_activity_logs_user_id ON activity_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_activity_logs_resource ON activity_logs(resource_type, resource_id);
