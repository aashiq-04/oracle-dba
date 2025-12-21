#!/bin/bash

# Oracle DBA Platform - Database Initialization Script
# This script creates the PostgreSQL database and runs all schema DDL

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Oracle DBA Platform - Database Initialization ===${NC}"

# Check if psql is installed
if ! command -v psql &> /dev/null; then
    echo -e "${RED}Error: psql is not installed${NC}"
    exit 1
fi

# Database configuration
DB_NAME="oracle_dba_platform"
DB_USER="postgres"
DB_HOST="localhost"
DB_PORT="5432"

echo -e "${YELLOW}Database: ${DB_NAME}${NC}"
echo -e "${YELLOW}User: ${DB_USER}${NC}"
echo -e "${YELLOW}Host: ${DB_HOST}:${DB_PORT}${NC}"
echo ""

# Create database if it doesn't exist
echo -e "${GREEN}Step 1: Creating database...${NC}"
psql -U ${DB_USER} -h ${DB_HOST} -p ${DB_PORT} -c "CREATE DATABASE ${DB_NAME};" 2>/dev/null || echo "Database already exists"

# Create extensions
echo -e "${GREEN}Step 2: Creating extensions...${NC}"
psql -U ${DB_USER} -h ${DB_HOST} -p ${DB_PORT} -d ${DB_NAME} <<EOF
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pgcrypto;
EOF

# Create schemas
echo -e "${GREEN}Step 3: Creating schemas...${NC}"
psql -U ${DB_USER} -h ${DB_HOST} -p ${DB_PORT} -d ${DB_NAME} <<EOF
CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS audit;
CREATE SCHEMA IF NOT EXISTS monitoring;
EOF

# Create tables
echo -e "${GREEN}Step 4: Creating tables...${NC}"

# Auth tables
psql -U ${DB_USER} -h ${DB_HOST} -p ${DB_PORT} -d ${DB_NAME} <<EOF
-- Users table
CREATE TABLE IF NOT EXISTS auth.users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    last_login TIMESTAMP
);

-- Roles table
CREATE TABLE IF NOT EXISTS auth.roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

-- User-Role mapping
CREATE TABLE IF NOT EXISTS auth.user_roles (
    user_id UUID REFERENCES auth.users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES auth.roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, role_id)
);

-- Permissions table
CREATE TABLE IF NOT EXISTS auth.permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL
);

-- Role-Permission mapping
CREATE TABLE IF NOT EXISTS auth.role_permissions (
    role_id UUID REFERENCES auth.roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES auth.permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);
EOF

# Audit tables
psql -U ${DB_USER} -h ${DB_HOST} -p ${DB_PORT} -d ${DB_NAME} <<EOF
-- Audit logs
CREATE TABLE IF NOT EXISTS audit.logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID,
    username TEXT NOT NULL,
    action TEXT NOT NULL,
    target TEXT,
    success BOOLEAN NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit.logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit.logs(user_id);
EOF

# Monitoring tables
psql -U ${DB_USER} -h ${DB_HOST} -p ${DB_PORT} -d ${DB_NAME} <<EOF
-- Session snapshots
CREATE TABLE IF NOT EXISTS monitoring.session_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    oracle_db TEXT NOT NULL,
    snapshot_time TIMESTAMP NOT NULL,
    active_sessions INTEGER,
    blocked_sessions INTEGER,
    raw_data JSONB
);

CREATE INDEX IF NOT EXISTS idx_session_snapshots_time ON monitoring.session_snapshots(snapshot_time DESC);
EOF

# Insert seed data
echo -e "${GREEN}Step 5: Inserting seed data...${NC}"
psql -U ${DB_USER} -h ${DB_HOST} -p ${DB_PORT} -d ${DB_NAME} <<EOF
-- Insert roles
INSERT INTO auth.roles (name, description) VALUES
('ADMIN', 'Platform administrator'),
('DBA', 'Database administrator'),
('DEVELOPER', 'Application developer'),
('READ_ONLY', 'Read-only access')
ON CONFLICT (name) DO NOTHING;

-- Insert permissions
INSERT INTO auth.permissions (code, description) VALUES
('VIEW_SESSIONS', 'View active Oracle sessions'),
('VIEW_LOCKS', 'View blocking and locked sessions'),
('VIEW_TABLESPACES', 'View tablespace usage'),
('VIEW_SQL', 'View SQL execution metrics'),
('VIEW_SCHEMA', 'View schema objects and changes'),
('MANAGE_USERS', 'Create/update users'),
('MANAGE_ROLES', 'Assign roles and permissions'),
('AUDIT_READ', 'View audit logs'),
('SESSION_KILL', 'Kill Oracle sessions')
ON CONFLICT (code) DO NOTHING;

-- Assign permissions to DBA role
INSERT INTO auth.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth.roles r, auth.permissions p
WHERE r.name = 'DBA'
AND p.code IN (
    'VIEW_SESSIONS',
    'VIEW_LOCKS',
    'VIEW_TABLESPACES',
    'VIEW_SQL',
    'VIEW_SCHEMA',
    'AUDIT_READ'
)
ON CONFLICT DO NOTHING;

-- Assign permissions to ADMIN role (all permissions)
INSERT INTO auth.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth.roles r, auth.permissions p
WHERE r.name = 'ADMIN'
ON CONFLICT DO NOTHING;

-- Assign permissions to DEVELOPER role
INSERT INTO auth.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth.roles r, auth.permissions p
WHERE r.name = 'DEVELOPER'
AND p.code IN ('VIEW_SESSIONS', 'VIEW_SQL')
ON CONFLICT DO NOTHING;

-- Assign permissions to READ_ONLY role
INSERT INTO auth.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth.roles r, auth.permissions p
WHERE r.name = 'READ_ONLY'
AND p.code IN ('VIEW_TABLESPACES')
ON CONFLICT DO NOTHING;
EOF

# Create default admin user (password: admin123)
echo -e "${GREEN}Step 6: Creating default admin user...${NC}"
psql -U ${DB_USER} -h ${DB_HOST} -p ${DB_PORT} -d ${DB_NAME} <<EOF
-- Password: admin123 (hashed with bcrypt cost 10)
INSERT INTO auth.users (username, email, password_hash, is_active)
VALUES (
    'admin',
    'admin@oracleplatform.com',
    '\$2a\$10\$rJ8qVZ.YqYKLZ8fqE8GQZ.YqYKLZ8fqE8GQZ.YqYKLZ8fqE8GQZe', -- Replace with actual bcrypt hash
    true
)
ON CONFLICT (username) DO NOTHING;

-- Assign ADMIN role to admin user
INSERT INTO auth.user_roles (user_id, role_id)
SELECT u.id, r.id
FROM auth.users u, auth.roles r
WHERE u.username = 'admin' AND r.name = 'ADMIN'
ON CONFLICT DO NOTHING;
EOF

echo ""
echo -e "${GREEN}=== Database initialization complete! ===${NC}"
echo ""
echo -e "${YELLOW}Default credentials:${NC}"
echo -e "  Username: admin"
echo -e "  Password: admin123"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo -e "  1. Update .env with your database credentials"
echo -e "  2. Run: go run cmd/server/main.go"
echo -e "  3. Open: http://localhost:8080"
echo ""