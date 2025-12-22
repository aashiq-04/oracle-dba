// Authentication Types
export interface User {
    id: string;
    username: string;
    email: string;
    roles: Role[];
    lastLogin?: string;
    createdAt: string;
  }
  
  export interface Role {
    id: string;
    name: string;
    description: string;
  }
  
  export interface AuthPayload {
    token: string;
    user: User;
    expiresAt: string;
  }
  
  // Session Types
  export interface OracleSession {
    sid: number;
    serial: number;
    username?: string;
    schemaName?: string;
    osUser?: string;
    machine?: string;
    program?: string;
    status: 'ACTIVE' | 'INACTIVE';
    sqlId?: string;
    sqlText?: string;
    logonTime?: string;
    lastCallSeconds: number;
    blockingSession?: number;
    waitClass?: string;
    event?: string;
    secondsInWait?: number;
  }
  
  export interface SessionSummary {
    totalSessions: number;
    activeSessions: number;
    inactiveSessions: number;
    blockedSessions: number;
    bySchema: SessionsBySchema[];
  }
  
  export interface SessionsBySchema {
    schemaName: string;
    total: number;
    active: number;
    inactive: number;
  }
  
  export interface BlockingSession {
    blockingSid: number;
    blockingSerial: number;
    blockingUser?: string;
    blockingSchema?: string;
    blockingStatus: string;
    blockingSqlId?: string;
    blockingSqlText?: string;
    blockedSid: number;
    blockedSerial: number;
    blockedUser?: string;
    blockedSchema?: string;
    blockedWaitClass?: string;
    blockedEvent?: string;
    blockedDurationSeconds: number;
    blockedSqlText?: string;
  }
  
  // Tablespace Types
  export interface Tablespace {
    name: string;
    totalSizeMb: number;
    usedSizeMb: number;
    freeSizeMb: number;
    usagePercentage: number;
    status: string;
    contents: 'PERMANENT' | 'TEMPORARY' | 'UNDO';
    datafileCount: number;
  }
  
  // SQL Performance Types
  export interface SqlPerformance {
    sqlId: string;
    sqlText?: string;
    schemaName?: string;
    parsingSchema?: string;
    executions: number;
    elapsedTimeMs: number;
    avgElapsedMs: number;
    cpuTimeMs: number;
    avgCpuMs: number;
    diskReads: number;
    bufferGets: number;
    rowsProcessed: number;
    firstLoadTime?: string;
    lastActiveTime?: string;
  }
  
  // Schema Types
  export interface SchemaInfo {
    schemaName: string;
    totalObjects: number;
    tableCount: number;
    indexCount: number;
    viewCount: number;
    procedureCount: number;
    functionCount: number;
    packageCount: number;
  }
  
  // Database Health Types
  export interface DatabaseInstance {
    instanceName: string;
    hostName: string;
    version: string;
    startupTime: string;
    status: string;
    databaseStatus: string;
    instanceRole: string;
    uptimeDays: number;
  }