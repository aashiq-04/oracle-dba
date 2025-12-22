import { gql } from '@apollo/client';

// Authentication Queries
export const LOGIN_MUTATION = gql`
  mutation Login($username: String!, $password: String!) {
    login(input: { username: $username, password: $password }) {
      token
      user {
        id
        username
        email
        roles {
          id
          name
          description
        }
      }
      expiresAt
    }
  }
`;

export const ME_QUERY = gql`
  query Me {
    me {
      id
      username
      email
      roles {
        id
        name
        description
      }
      lastLogin
      createdAt
    }
  }
`;

// Session Queries
export const ACTIVE_SESSIONS_QUERY = gql`
  query ActiveSessions {
    activeSessions {
      sid
      serial
      username
      schemaName
      osUser
      machine
      program
      status
      sqlId
      sqlText
      logonTime
      lastCallSeconds
      blockingSession
      waitClass
      event
      secondsInWait
    }
  }
`;

export const SESSION_SUMMARY_QUERY = gql`
  query SessionSummary {
    sessionSummary {
      totalSessions
      activeSessions
      inactiveSessions
      blockedSessions
      bySchema {
        schemaName
        total
        active
        inactive
      }
    }
  }
`;

export const BLOCKING_SESSIONS_QUERY = gql`
  query BlockingSessions {
    blockingSessions {
      blockingSid
      blockingSerial
      blockingUser
      blockingSchema
      blockingStatus
      blockingSqlId
      blockingSqlText
      blockedSid
      blockedSerial
      blockedUser
      blockedSchema
      blockedWaitClass
      blockedEvent
      blockedDurationSeconds
      blockedSqlText
    }
  }
`;

// Tablespace Queries
export const TABLESPACES_QUERY = gql`
  query Tablespaces {
    tablespaces {
      name
      totalSizeMb
      usedSizeMb
      freeSizeMb
      usagePercentage
      status
      contents
      datafileCount
    }
  }
`;

// SQL Performance Queries
export const TOP_SQL_BY_ELAPSED_QUERY = gql`
  query TopSqlByElapsedTime($limit: Int!) {
    topSqlByElapsedTime(limit: $limit) {
      sqlId
      sqlText
      schemaName
      parsingSchema
      executions
      elapsedTimeMs
      avgElapsedMs
      cpuTimeMs
      avgCpuMs
      diskReads
      bufferGets
      rowsProcessed
      firstLoadTime
      lastActiveTime
    }
  }
`;

export const TOP_SQL_BY_CPU_QUERY = gql`
  query TopSqlByCpuTime($limit: Int!) {
    topSqlByCpuTime(limit: $limit) {
      sqlId
      sqlText
      schemaName
      parsingSchema
      executions
      elapsedTimeMs
      avgElapsedMs
      cpuTimeMs
      avgCpuMs
      diskReads
      bufferGets
      rowsProcessed
    }
  }
`;

// Schema Queries
export const SCHEMAS_QUERY = gql`
  query Schemas {
    schemas {
      schemaName
      totalObjects
      tableCount
      indexCount
      viewCount
      procedureCount
      functionCount
      packageCount
    }
  }
`;

// Database Health Queries
export const DATABASE_INSTANCE_QUERY = gql`
  query DatabaseInstance {
    databaseInstance {
      instanceName
      hostName
      version
      startupTime
      status
      databaseStatus
      instanceRole
      uptimeDays
    }
  }
`;

// Logout Mutation
export const LOGOUT_MUTATION = gql`
  mutation Logout {
    logout
  }
`;