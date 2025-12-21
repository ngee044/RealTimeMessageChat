# DBWorker - Database Integration for RealTimeMessageChat

## Overview

DBWorker is a specialized background job component that handles asynchronous database operations for storing encrypted broadcast messages in RealTimeMessageChat. It integrates seamlessly with the MainServerConsumer to persist messages consumed from RabbitMQ into PostgreSQL.

## Architecture

```
RabbitMQ → MainServerConsumer → [Redis + DBWorker] → PostgreSQL
                                      ↓
                                  MainServer → TCP Clients
```

**Message Flow:**
1. Message arrives in RabbitMQ queue
2. MainServerConsumer validates and consumes the message
3. Message is stored in Redis for immediate broadcast
4. DBWorker job is pushed to low-priority thread pool
5. DBWorker encrypts message (if enabled) and stores in PostgreSQL asynchronously
6. MainServer reads from Redis and broadcasts to connected clients

## Features

- **Asynchronous Storage**: Non-blocking database writes using thread pool
- **Optional Encryption**: AES-256-CBC encryption before storage
- **Fault Tolerant**: Database errors don't block message broadcasting
- **Thread-Safe**: Runs in isolated low-priority thread workers
- **JSON Validation**: Strict message structure validation
- **SQL Injection Protection**: Parameterized queries with input escaping

## Database Schema

### Messages Table

```sql
CREATE TABLE messages (
    message_id BIGSERIAL PRIMARY KEY,
    id VARCHAR(255) NOT NULL,           -- User identifier
    sub_id VARCHAR(255) NOT NULL,       -- Session identifier
    publisher_info TEXT,                 -- JSON publisher metadata
    server_name VARCHAR(100) NOT NULL,   -- Target server (default: MainServer)
    message_content TEXT NOT NULL,       -- Encrypted or plain message
    is_encrypted BOOLEAN NOT NULL,       -- Encryption flag
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

**Indexes:**
- `idx_messages_id` - User ID lookups
- `idx_messages_created_at` - Chronological queries
- `idx_messages_id_created_at` - Composite user timeline queries

### Encryption Keys Table (Optional)

Stores encryption keys for key rotation and management.

```sql
CREATE TABLE encryption_keys (
    key_id SERIAL PRIMARY KEY,
    key_name VARCHAR(100) UNIQUE NOT NULL,
    encryption_key TEXT NOT NULL,        -- Base64-encoded AES key
    encryption_iv TEXT NOT NULL,          -- Base64-encoded IV
    is_active BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    rotated_at TIMESTAMP WITH TIME ZONE
);
```

## Configuration

### MainServerConsumer Configuration

Add to `main_server_consumer_configurations.json`:

```json
{
  "use_database": true,
  "database_connection_string": "host=postgres port=5432 dbname=rtmc user=rtmc_user password=rtmc_password",
  "database_encryption_enabled": false,
  "database_encryption_key": "",
  "database_encryption_iv": ""
}
```

#### Configuration Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `use_database` | boolean | `false` | Enable/disable database storage |
| `database_connection_string` | string | `""` | PostgreSQL connection string (libpq format) |
| `database_encryption_enabled` | boolean | `false` | Enable message encryption before storage |
| `database_encryption_key` | string | `""` | Base64-encoded AES-256 encryption key |
| `database_encryption_iv` | string | `""` | Base64-encoded initialization vector |

### Generating Encryption Keys

Use the provided utility to generate secure encryption keys:

```bash
cd scripts
g++ -std=c++20 -DUSE_ENCRYPT_MODULE \
    -I../.cpp_tool_kit/Utilities \
    generate_encryption_keys.cpp \
    ../.cpp_tool_kit/Utilities/Encryptor.cpp \
    -lcryptopp -o generate_encryption_keys

./generate_encryption_keys
```

**Output:**
```json
"database_encryption_enabled": true,
"database_encryption_key": "aGVsbG8gd29ybGQgdGhpcyBpcyBhIHRlc3Q=",
"database_encryption_iv": "MTIzNDU2Nzg5MGFiY2RlZg=="
```

**Security Notes:**
- Store keys in environment variables or secrets manager (not in version control)
- Rotate keys periodically
- Keep secure backups of encryption keys
- Lost keys = unrecoverable encrypted data

## Message Format

### Expected JSON Structure

```json
{
  "id": "user_123",
  "sub_id": "session_456",
  "publisher_information": {
    "client_ip": "192.168.1.100",
    "timestamp": "2025-09-29T10:30:00Z"
  },
  "message": {
    "server_name": "MainServer",
    "content": "Hello, world!"
  }
}
```

### Field Validation

| Field | Required | Type | Description |
|-------|----------|------|-------------|
| `id` | Yes | String | User identifier |
| `sub_id` | Yes | String | Session/subscription ID |
| `publisher_information` | No | Object | Publisher metadata (default: `{}`) |
| `message` | Yes | Object | Message payload |
| `message.server_name` | No | String | Target server (default: `"MainServer"`) |
| `message.content` | Yes | String | Message content |

## Database Setup

### Local Development

1. **Install PostgreSQL:**
   ```bash
   # macOS
   brew install postgresql@16
   brew services start postgresql@16

   # Ubuntu/Debian
   sudo apt-get install postgresql-16
   sudo systemctl start postgresql
   ```

2. **Create Database:**
   ```bash
   psql -U postgres
   ```
   ```sql
   CREATE DATABASE rtmc;
   CREATE USER rtmc_user WITH PASSWORD 'rtmc_password';
   GRANT ALL PRIVILEGES ON DATABASE rtmc TO rtmc_user;
   \c rtmc
   \i database/schema.sql
   ```

3. **Update Configuration:**
   ```json
   {
     "use_database": true,
     "database_connection_string": "host=localhost port=5432 dbname=rtmc user=rtmc_user password=rtmc_password"
   }
   ```

### Docker Deployment

The docker-compose configuration automatically sets up PostgreSQL:

```bash
cd docker
./docker-compose.sh
```

**Services:**
- PostgreSQL: `postgres:5432`
- Schema auto-initialized from `database/schema.sql`
- Persistent volume: `postgres-data`

**Verify Database:**
```bash
docker exec -it postgres psql -U rtmc_user -d rtmc

# Inside psql:
\dt                                    # List tables
SELECT COUNT(*) FROM messages;         # Check message count
SELECT * FROM recent_messages LIMIT 5; # View recent messages
```

## Usage Examples

### Check Database Connection

```bash
docker logs mainserver-consumer | grep -i database
```

**Expected Output:**
```
[INFO] Database client initialized successfully
[INFO] DBWorker job pushed to thread pool
[INFO] DBWorker: Message stored successfully (id: user_123, sub_id: session_456, encrypted: false)
```

### Query Messages

```sql
-- Get all messages for a user
SELECT * FROM messages
WHERE id = 'user_123'
ORDER BY created_at DESC;

-- Get encrypted messages
SELECT message_id, id, created_at
FROM messages
WHERE is_encrypted = TRUE;

-- Daily message statistics
SELECT * FROM message_statistics
WHERE message_date >= CURRENT_DATE - INTERVAL '7 days';

-- Clean up old messages (older than 30 days)
SELECT cleanup_old_messages(30);
```

### Monitor Performance

```sql
-- Table size
SELECT pg_size_pretty(pg_total_relation_size('messages'));

-- Message count by server
SELECT server_name, COUNT(*),
       COUNT(*) FILTER (WHERE is_encrypted) as encrypted_count
FROM messages
GROUP BY server_name;

-- Recent activity
SELECT DATE(created_at) as date, COUNT(*) as count
FROM messages
WHERE created_at >= NOW() - INTERVAL '7 days'
GROUP BY DATE(created_at)
ORDER BY date DESC;
```

## Error Handling

### Database Errors

DBWorker handles errors gracefully without blocking message broadcasting:

**Common Errors:**
1. **Connection Failed**: Database unreachable
   - **Impact**: Message not stored in DB, but still broadcast via Redis
   - **Log**: `Failed to initialize database: <error>`
   - **Action**: Check connection string, verify PostgreSQL is running

2. **SQL Syntax Error**: Invalid query
   - **Impact**: Message not stored
   - **Log**: `DBWorker: Failed to store message to database`
   - **Action**: Check schema version, verify table exists

3. **Encryption Error**: Invalid key/IV
   - **Impact**: Message stored unencrypted with warning
   - **Log**: `Encryption failed, storing plain text`
   - **Action**: Verify encryption keys are valid base64

### Debugging

Enable detailed logging:

```json
{
  "write_console_log": 3,  // Information level
  "write_file_log": 3      // Log to file
}
```

**Log Locations:**
- Docker: `/opt/app/logs/` (mounted to `docker/logs/consumer/`)
- Local: Configured via `log_root_path`

## Performance Tuning

### Thread Pool Configuration

Adjust low-priority worker count for database operations:

```json
{
  "high_priority_count": 3,   // Connection handling
  "normal_priority_count": 3, // Message processing
  "low_priority_count": 5     // Database writes (increase for high volume)
}
```

**Recommendations:**
- Low volume (<100 msg/s): `low_priority_count: 3-5`
- Medium volume (100-500 msg/s): `low_priority_count: 10-15`
- High volume (>500 msg/s): `low_priority_count: 20+`, consider sharding

### Database Optimization

1. **Indexes**: Already optimized for common queries
2. **Connection Pooling**: Use pgBouncer for production
3. **Partitioning**: Partition `messages` by date for very large datasets
4. **Archive Old Messages**: Use `cleanup_old_messages()` function regularly

```sql
-- Set up automatic cleanup (cron job)
-- Keep only last 90 days
SELECT cleanup_old_messages(90);
```

## Security Best Practices

1. **Connection String**: Use environment variables
   ```bash
   export DB_CONN="host=postgres port=5432 dbname=rtmc user=rtmc_user password=rtmc_password"
   ```

2. **Encryption Keys**: Store in secrets manager (AWS Secrets, Vault, etc.)

3. **Network Security**: Restrict PostgreSQL access via firewall
   ```
   # PostgreSQL pg_hba.conf
   host rtmc rtmc_user 10.0.0.0/8 scram-sha-256
   ```

4. **SSL/TLS**: Enable PostgreSQL SSL for production
   ```
   "database_connection_string": "host=postgres sslmode=require ..."
   ```

5. **Principle of Least Privilege**: Grant minimal permissions
   ```sql
   REVOKE ALL ON DATABASE rtmc FROM PUBLIC;
   GRANT CONNECT ON DATABASE rtmc TO rtmc_user;
   GRANT SELECT, INSERT ON messages TO rtmc_user;
   ```

## Testing

### Unit Tests

Run DBWorker tests:

```bash
cd build
ctest -R DBWorker -V
```

### Integration Testing

1. **Send Test Message:**
   ```bash
   # Using RabbitMQ management UI or CLI
   rabbitmqadmin publish routing_key=message_broadcast_queue \
     payload='{"id":"test_user","sub_id":"test_session","message":{"content":"test"}}'
   ```

2. **Verify Storage:**
   ```bash
   docker exec -it postgres psql -U rtmc_user -d rtmc \
     -c "SELECT * FROM messages WHERE id='test_user' ORDER BY created_at DESC LIMIT 1;"
   ```

3. **Check Logs:**
   ```bash
   docker logs mainserver-consumer | tail -n 20
   ```

## Troubleshooting

### Database Not Storing Messages

**Check:**
1. `use_database` is `true` in config
2. PostgreSQL is running and accessible
3. Connection string is correct
4. Schema is initialized (`\dt` in psql)
5. Logs show database initialization success

### Encryption Not Working

**Check:**
1. `USE_ENCRYPT_MODULE` is defined during compilation
2. Encryption keys are valid base64 strings
3. Keys are not empty when `database_encryption_enabled: true`
4. CryptoPP library is linked correctly

### Performance Issues

**Check:**
1. Increase `low_priority_count`
2. Monitor thread pool queue depth
3. Check PostgreSQL connection pooling
4. Review table indexes (`\d messages`)
5. Consider database sharding for very high volume

## Maintenance

### Regular Tasks

1. **Backup Database:**
   ```bash
   docker exec postgres pg_dump -U rtmc_user rtmc > backup_$(date +%Y%m%d).sql
   ```

2. **Monitor Disk Usage:**
   ```sql
   SELECT pg_size_pretty(pg_database_size('rtmc'));
   ```

3. **Clean Old Messages:**
   ```sql
   SELECT cleanup_old_messages(30); -- Keep last 30 days
   ```

4. **Rotate Encryption Keys** (if enabled):
   ```sql
   -- Mark old key as inactive
   UPDATE encryption_keys SET is_active = FALSE WHERE key_name = 'old_key';
   -- Insert new key
   INSERT INTO encryption_keys (key_name, encryption_key, encryption_iv, is_active)
   VALUES ('new_key', '<new_key_base64>', '<new_iv_base64>', TRUE);
   ```

## Future Enhancements

- [ ] Batch inserts for higher throughput
- [ ] Database connection pooling
- [ ] Automatic key rotation
- [ ] Message archival to S3/cold storage
- [ ] Full-text search on message content
- [ ] Analytics and reporting views
- [ ] Multi-database support (MySQL, MongoDB)
- [ ] Distributed tracing integration

## Support

For issues or questions:
1. Check logs: `docker logs mainserver-consumer`
2. Verify configuration: `cat docker/config/main_server_consumer_configurations.json`
3. Test database: `docker exec -it postgres psql -U rtmc_user -d rtmc`
4. Review documentation: `CLAUDE.md` in project root

## License

Part of RealTimeMessageChat project. See main project LICENSE.
