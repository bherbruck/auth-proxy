# Auth Proxy

`auth-proxy` is a simple single-executable authentication proxy service similar to oauth2-proxy but designed for a single user. It provides session-based authentication with a beautiful login interface built using franken-ui.

## Features

- 🔒 Single-user authentication with session management
- 🎨 Beautiful, responsive login interface using franken-ui
- 🔄 Reverse proxy functionality to protect any application
- 🔐 Support for both plain text and bcrypt hashed passwords
- ⚙️ Configurable via environment variables or command line arguments
- 📱 Mobile-friendly login page with gradient background
- 🚀 Single binary deployment - no dependencies
- 🔑 Automatic cookie secret generation if not provided

## Quick Start

1. **Build the application:**
   ```bash
   go build -o auth-proxy ./cmd/authproxy
   ```

2. **Run with basic configuration:**
   ```bash
   AUTH_PROXY_USERNAME=admin \
   AUTH_PROXY_PASSWORD=your-password \
   AUTH_PROXY_TARGET=http://localhost:3000 \
   ./auth-proxy
   ```

3. **Access your application:**
   - Navigate to `http://localhost:8080`
   - Login with your credentials
   - Access your protected application seamlessly

## Configuration

### Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `AUTH_PROXY_USERNAME` | Username for authentication | `admin` |
| `AUTH_PROXY_PASSWORD` | Password for authentication (use this OR `AUTH_PROXY_PASSWORD_HASH`) | `my-secure-password` |
| `AUTH_PROXY_TARGET` | Target application URL to proxy to | `http://localhost:3000` |

### Optional Environment Variables

| Variable | Default | Description | Example |
|----------|---------|-------------|---------|
| `AUTH_PROXY_COOKIE_SECRET` | auto-generated | Secret key for cookie encryption | `your-32-character-secret-key` |
| `AUTH_PROXY_LOGIN_TITLE` | `Auth Proxy` | Custom title for the login page | `My Application` |
| `AUTH_PROXY_PORT` | `8080` | Port to run the auth proxy on | `9000` |
| `AUTH_PROXY_PASSWORD_HASH` | - | Bcrypt hash of password (replaces `AUTH_PROXY_PASSWORD`) | `$2a$10$...` |

### Command Line Arguments

You can also use command line arguments instead of environment variables:

```bash
./auth-proxy \
  --auth-proxy-username=admin \
  --auth-proxy-password=my-password \
  --auth-proxy-target=http://localhost:3000 \
  --auth-proxy-login-title="My App"
```

## Usage Examples

### Basic Setup with Environment File

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` with your configuration:
   ```bash
   AUTH_PROXY_USERNAME=admin
   AUTH_PROXY_PASSWORD=your-secure-password
   AUTH_PROXY_TARGET=http://localhost:3000
   AUTH_PROXY_LOGIN_TITLE=My Application
   ```

3. Run with environment file:
   ```bash
   source .env && ./auth-proxy
   ```

### Using Bcrypt Password Hash

For enhanced security, use a bcrypt hash instead of plain text password:

```bash
# Generate bcrypt hash (you can use online tools or bcrypt CLI)
echo "my-password" | bcrypt

# Use the hash in your configuration
AUTH_PROXY_USERNAME=admin \
AUTH_PROXY_PASSWORD_HASH='$2a$10$...' \
AUTH_PROXY_TARGET=http://localhost:3000 \
./auth-proxy
```

### Docker Usage

#### Quick Start with Docker Compose

```bash
# Build and start all services
docker compose up --build

# Run in background
docker compose up -d --build

# View logs
docker compose logs -f auth-proxy

# Stop services
docker compose down
```

#### Custom Configuration

Create a `.env` file for custom configuration:

```bash
# .env file
AUTH_PROXY_USERNAME=admin
AUTH_PROXY_PASSWORD=your-secure-password
AUTH_PROXY_LOGIN_TITLE=My Protected App
AUTH_PROXY_COOKIE_SECRET=your-32-character-secret-key-here
```

Then run:
```bash
docker compose up --build
```

#### Manual Docker Build

```bash
# Build the image
docker build -t auth-proxy .

# Run with custom target
docker run -p 8080:8080 \
  -e AUTH_PROXY_USERNAME=admin \
  -e AUTH_PROXY_PASSWORD=secure-password \
  -e AUTH_PROXY_TARGET=http://host.docker.internal:3000 \
  auth-proxy
```

#### Docker Features

- **Multi-stage build**: Optimized for minimal final image size
- **Distroless base**: Secure, minimal runtime with CA certificates
- **Health checks**: Built-in container health monitoring  
- **Non-root user**: Runs as unprivileged user for security
- **Resource limits**: CPU and memory constraints configured
- **Service networking**: Isolated Docker network for inter-service communication

## How It Works

1. **Authentication Flow:**
   - User accesses any protected URL
   - If not authenticated, redirected to `/auth/login`
   - User enters credentials on the beautiful franken-ui login page
   - On successful login, session cookie is created
   - User is redirected to originally requested URL

2. **Proxy Flow:**
   - All authenticated requests are proxied to the target application
   - Headers, query parameters, and body are forwarded unchanged
   - Responses from target application are returned to the user

3. **Session Management:**
   - Secure HTTP-only cookies with 7-day expiration
   - Sessions automatically invalidated on logout
   - Cookie secret ensures session security

## API Endpoints

- `GET /auth/login` - Login page
- `POST /auth/login` - Process login credentials
- `GET|POST /auth/logout` - Logout and clear session
- `/*` - All other paths are proxied to target application (requires authentication)

## Frontend

The login interface features:

- 🎨 Modern, responsive design using [franken-ui](https://franken-ui.dev/docs/2.1)
- 🌈 Beautiful gradient background
- 📱 Mobile-friendly responsive layout
- ⚡ Fast loading with CDN-hosted assets
- 🎯 Centered login card with clean styling
- 🔒 Proper form validation and error handling
- ✨ Smooth animations and transitions

## Security Features

- Session-based authentication with secure cookies
- HTTP-only cookies prevent XSS attacks
- Configurable cookie secret for session encryption
- Support for bcrypt password hashing
- Automatic redirect to intended page after login
- Clean session invalidation on logout
- CSRF protection through proper form handling

## Development

### Testing

A test server is included for development:

```bash
# Terminal 1: Start test target server
go run test-server.go

# Terminal 2: Start auth proxy
AUTH_PROXY_USERNAME=admin \
AUTH_PROXY_PASSWORD=testpass123 \
AUTH_PROXY_TARGET=http://localhost:3000 \
./auth-proxy
```

Then visit `http://localhost:8080` to test the authentication flow.

### Building

```bash
# Build for current platform
go build -o auth-proxy ./cmd/authproxy

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o auth-proxy-linux ./cmd/authproxy

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o auth-proxy.exe ./cmd/authproxy
```

## Contributing

This project is designed to be simple and focused. Contributions are welcome for:

- Bug fixes
- Security improvements
- Documentation enhancements
- Performance optimizations

## License

MIT License - see LICENSE file for details.
