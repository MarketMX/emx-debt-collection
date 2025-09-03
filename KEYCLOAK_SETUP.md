# Keycloak Authentication Setup

This document describes the Keycloak authentication setup for the debt collection application.

## Architecture

- **Keycloak**: Latest version running on port 8081
- **Database**: PostgreSQL 15 (shared with main application)
- **Network**: Custom Docker bridge network for service communication
- **Persistence**: Named volumes for both PostgreSQL and Keycloak data

## Environment Variables

The following environment variables have been added to `.env`:

### Keycloak Configuration
- `KEYCLOAK_PORT=8081` - Port for Keycloak admin console and API
- `KEYCLOAK_ADMIN_USER=admin` - Admin username for Keycloak console
- `KEYCLOAK_ADMIN_PASSWORD=admin123` - Admin password (change in production!)

### Keycloak Database Configuration
- `KEYCLOAK_DB_HOST=psql_bp` - PostgreSQL service name
- `KEYCLOAK_DB_PORT=5432` - PostgreSQL port
- `KEYCLOAK_DB_DATABASE=keycloak` - Keycloak database name
- `KEYCLOAK_DB_USERNAME=keycloak` - Keycloak database user
- `KEYCLOAK_DB_PASSWORD=keycloak123` - Keycloak database password

## Services

### PostgreSQL Service (`psql_bp`)
- **Image**: `postgres:15-alpine`
- **Databases**: 
  - `blueprint` (main application)
  - `keycloak` (authentication service)
- **Initialization**: Automatic Keycloak database and user creation via `/docker/postgresql/init-keycloak-db.sh`
- **Health Check**: PostgreSQL ready check with 30s startup period
- **Persistence**: `psql_volume_bp` named volume

### Keycloak Service (`keycloak`)
- **Image**: `quay.io/keycloak/keycloak:latest`
- **Mode**: Development mode (`start-dev`)
- **Admin Console**: `http://localhost:8081`
- **Database**: PostgreSQL backend
- **Health Check**: HTTP health endpoint with 60s startup period
- **Persistence**: `keycloak_data` named volume
- **Features Enabled**:
  - Health checks (`KC_HEALTH_ENABLED=true`)
  - Metrics (`KC_METRICS_ENABLED=true`)
  - HTTP (non-HTTPS for development)

## Database Schema

The PostgreSQL initialization script creates:
- Keycloak database with proper ownership
- Keycloak user with full privileges
- Schema permissions for Keycloak operations

## Starting the Services

```bash
# Start all services
docker-compose up -d

# Start only Keycloak and dependencies
docker-compose up -d keycloak

# View logs
docker-compose logs -f keycloak
docker-compose logs -f psql_bp
```

## Accessing Keycloak

1. **Admin Console**: http://localhost:8081
2. **Credentials**: 
   - Username: `admin`
   - Password: `admin123`

## Health Checks

- **PostgreSQL**: `pg_isready` check every 10 seconds
- **Keycloak**: HTTP health endpoint check every 30 seconds

## Security Considerations

### Development vs Production

Current setup is optimized for **local development**:

- HTTP enabled (no HTTPS)
- Hostname strictness disabled
- Simple passwords in environment variables

### Production Recommendations

1. **Enable HTTPS**:
   ```yaml
   KC_HTTP_ENABLED: false
   KC_HTTPS_CERTIFICATE_FILE: /path/to/cert.pem
   KC_HTTPS_CERTIFICATE_KEY_FILE: /path/to/key.pem
   ```

2. **Use Docker Secrets**:
   ```yaml
   secrets:
     - keycloak_admin_password
     - keycloak_db_password
   ```

3. **Strong Passwords**: Generate secure passwords for all accounts

4. **Network Security**: Use internal networks for database communication

5. **Resource Limits**: Add memory and CPU limits

## Integration with Application

### JWT Token Configuration

1. **Create Realm**: Set up a realm for your application
2. **Create Client**: Configure OAuth2/OpenID Connect client
3. **User Management**: Set up user roles and permissions
4. **Token Validation**: Configure your application to validate JWT tokens

### Environment Variables for Application

Add these to your application's environment:
```bash
KEYCLOAK_URL=http://keycloak:8080
KEYCLOAK_REALM=your-realm
KEYCLOAK_CLIENT_ID=your-client-id
KEYCLOAK_CLIENT_SECRET=your-client-secret
```

## Troubleshooting

### Common Issues

1. **Keycloak fails to start**: Check database connection and credentials
2. **Database connection errors**: Verify PostgreSQL is healthy before Keycloak starts
3. **Port conflicts**: Ensure ports 8081 and 5432 are available

### Useful Commands

```bash
# Check service health
docker-compose ps

# View service logs
docker-compose logs keycloak
docker-compose logs psql_bp

# Connect to PostgreSQL
docker-compose exec psql_bp psql -U melkey -d blueprint
docker-compose exec psql_bp psql -U keycloak -d keycloak

# Restart services
docker-compose restart keycloak
docker-compose restart psql_bp

# Clean up and rebuild
docker-compose down -v
docker-compose up -d
```

## Volume Management

- `psql_volume_bp`: PostgreSQL data persistence
- `keycloak_data`: Keycloak configuration and user data

To reset Keycloak data:
```bash
docker-compose down
docker volume rm emx-debt-collection_keycloak_data
docker-compose up -d keycloak
```

## Next Steps

1. **Realm Configuration**: Create application-specific realm
2. **Client Setup**: Configure OAuth2/OIDC client for your app
3. **User Management**: Set up user roles and groups
4. **Application Integration**: Update your Go application to validate JWT tokens
5. **Frontend Integration**: Implement login/logout flows