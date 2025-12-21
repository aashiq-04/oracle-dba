# Oracle DBA Platform

Enterprise-grade Oracle Database Monitoring, Management & Security Platform with strong emphasis on Oracle DBA concepts combined with backend engineering.

## 🎯 Project Goals

- Demonstrate deep understanding of Oracle Database internals
- Showcase backend/platform engineering skills
- Build production-ready enterprise architecture
- Strong focus on security, RBAC, and audit logging

## 🏗️ Architecture

```
┌─────────────────────────────────────────┐
│         GraphQL API Layer               │
│  (Authentication + RBAC Enforcement)    │
└──────────────┬──────────────────────────┘
               │
┌──────────────┴──────────────────────────┐
│         Service Layer                   │
│  - Auth Service                         │
│  - RBAC Service                         │
│  - Oracle Monitoring Service            │
└──────────────┬──────────────────────────┘
               │
┌──────────────┴──────────────────────────┐
│      Repository Layer                   │
│  - PostgreSQL (Users, Roles, Audit)     │
│  - Oracle Adapter (V$, DBA_ views)      │
└─────────────────────────────────────────┘
```

## 🚀 Features

### Core Monitoring Capabilities
- **Session Monitoring**: Track active/inactive sessions, blocking sessions
- **Lock Detection**: Identify blocking chains and lock contention
- **Tablespace Monitoring**: Space usage, growth trends
- **SQL Performance**: Slow queries, CPU usage, disk I/O analysis
- **Schema Monitoring**: Object counts, invalid objects, DDL changes
- **Database Health**: Instance info, uptime, version

### Security & RBAC
- **JWT Authentication**: Secure token-based auth
- **Role-Based Access Control**: Admin, DBA, Developer, Read-Only
- **Permission System**: Granular operation-level permissions
- **Audit Logging**: Comprehensive activity tracking

## 📋 Prerequisites

- **Go**: 1.21 or higher
- **PostgreSQL**: 13 or higher
- **Oracle Database**: 11g or higher (with DBA privileges)
- **Oracle Client**: Required for godror driver

## 🔧 Installation

### 1. Clone Repository

```bash
git clone <your-repo>
cd oracle-dba-platform
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configure Environment

Copy `.env.example` to `.env` and update values:

```bash
cp .env .env.local
```

Edit `.env`:
```env
# PostgreSQL (Platform Database)
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_password
POSTGRES_DB=oracle_dba_platform

# Oracle (Target Database to Monitor)
ORACLE_HOST=localhost
ORACLE_PORT=1521
ORACLE_SERVICE_NAME=ORCLPDB1
ORACLE_USERNAME=oramonitor
ORACLE_PASSWORD=your_oracle_password

# JWT Secret (generate with: openssl rand -base64 32)
JWT_SECRET=your_very_long_secret_key_at_least_32_characters
```

### 4. Initialize Database

```bash
# Run initialization script
./scripts/postgres/init_db.sh

# Or manually:
psql -U postgres -c "CREATE DATABASE oracle_dba_platform;"
psql -U postgres -d oracle_dba_platform -f scripts/postgres/schema.sql
```

### 5. Create Admin User

```bash
go run scripts/create_admin.go admin admin@example.com admin123
```

### 6. Setup Oracle Monitoring User

Connect to Oracle as SYSDBA and run:

```sql
-- Create monitoring user
CREATE USER oramonitor IDENTIFIED BY "SecurePassword123!";

-- Grant necessary privileges
GRANT CREATE SESSION TO oramonitor;
GRANT SELECT_CATALOG_ROLE TO oramonitor;
GRANT SELECT ON V$SESSION TO oramonitor;
GRANT SELECT ON V$SQL TO oramonitor;
GRANT SELECT ON V$LOCK TO oramonitor;
GRANT SELECT ON DBA_TABLESPACES TO oramonitor;
GRANT SELECT ON DBA_DATA_FILES TO oramonitor;
GRANT SELECT ON DBA_FREE_SPACE TO oramonitor;
GRANT SELECT ON DBA_OBJECTS TO oramonitor;
GRANT SELECT ON V$INSTANCE TO oramonitor;
```

## 🏃 Running the Application

### Development Mode

```bash
go run cmd/server/main.go
```

### Production Build

```bash
go build -o bin/oracle-dba-platform cmd/server/main.go
./bin/oracle-dba-platform
```

The server will start on `http://localhost:8080`

## 🧪 Testing the API

### 1. Access GraphQL Playground

Open browser: `http://localhost:8080`

### 2. Login

```graphql
mutation {
  login(input: {
    username: "admin"
    password: "admin123"
  }) {
    token
    user {
      id
      username
      email
      roles {
        name
      }
    }
    expiresAt
  }
}
```

### 3. Query Sessions (with token)

Add to HTTP Headers:
```json
{
  "Authorization": "Bearer YOUR_TOKEN_HERE"
}
```

Query:
```graphql
query {
  activeSessions {
    sid
    serial
    username
    schemaName
    status
    sqlText
    lastCallSeconds
  }
}
```

### 4. Monitor Blocking Sessions

```graphql
query {
  blockingSessions {
    blockingSid
    blockingUser
    blockedSid
    blockedUser
    blockedDurationSeconds
  }
}
```

### 5. Check Tablespace Usage

```graphql
query {
  tablespaces {
    name
    totalSizeMb
    usedSizeMb
    usagePercentage
    status
  }
}
```

## 📊 API Endpoints

- **GraphQL API**: `http://localhost:8080/query`
- **Playground**: `http://localhost:8080/`
- **Health Check**: `http://localhost:8080/health`

## 🔐 Default Roles & Permissions

### ADMIN
- Full system access
- User management
- Role assignment
- All monitoring capabilities

### DBA
- Session monitoring
- Lock detection
- Tablespace monitoring
- SQL performance analysis
- Schema monitoring
- Audit log access

### DEVELOPER
- Session monitoring (limited)
- SQL performance analysis

### READ_ONLY
- Tablespace monitoring only

## 📁 Project Structure

```
oracle-dba-platform/
├── cmd/
│   └── server/           # Main application entry point
├── internal/
│   ├── config/           # Configuration management
│   ├── database/         # Database connection pools
│   ├── graph/            # GraphQL schema & resolvers
│   ├── middleware/       # Auth, RBAC, logging
│   ├── repository/       # Database access layer
│   └── service/          # Business logic
├── pkg/
│   ├── logger/           # Structured logging
│   └── oracle/           # Oracle DB adapter
├── scripts/              # Database initialization scripts
├── .env                  # Environment configuration
├── go.mod
└── README.md
```

## 🛠️ Development

### Generate GraphQL Code

```bash
go run github.com/99designs/gqlgen generate
```

### Run Tests

```bash
go test ./...
```

### Format Code

```bash
go fmt ./...
```

## 📝 Key Design Decisions

1. **Backend-Only**: No frontend UI - focuses on API excellence
2. **GraphQL over REST**: Fine-grained queries, strong typing
3. **JWT Authentication**: Stateless, scalable
4. **RBAC at API Layer**: Every resolver checks permissions
5. **Read-Only Oracle Access**: Monitoring doesn't modify target DB
6. **Comprehensive Auditing**: All operations logged
7. **Clean Architecture**: Repository → Service → Resolver layers

## 🎓 Oracle DBA Concepts Demonstrated

- V$ dynamic performance views
- DBA_* data dictionary views
- Session management and monitoring
- Lock contention analysis
- Tablespace capacity planning
- SQL performance tuning
- Schema change tracking
- Database health monitoring


## 📧 Contact

Your Name - mohd04aashiq@gmail.com

## 📄 License

MIT License