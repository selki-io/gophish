#!/bin/bash
set -euo pipefail

# Script to set up Gophish on the VM
# This script is idempotent and can be run multiple times

LOG_FILE="/var/log/gophish-setup.log"
GOPHISH_USER="gophish"
GOPHISH_HOME="/opt/gophish"
GO_VERSION="1.20"
GOPHISH_BINARY_PATH="/opt/gophish/gophish"

# Function to log messages
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check Go version
check_go_version() {
    if command_exists go; then
        current_version=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+' || echo "0.0")
        required_version="$1"
        if [ "$(printf '%s\n' "$required_version" "$current_version" | sort -V | head -n1)" = "$required_version" ]; then
            return 0
        fi
    fi
    return 1
}

# Function to generate SSL certificates
generate_ssl_certificate() {
    local cert_name="$1"
    local key_path="$GOPHISH_HOME/ssl/${cert_name}.key"
    local cert_path="$GOPHISH_HOME/ssl/${cert_name}.crt"
    
    openssl req -x509 -newkey rsa:4096 -nodes \
        -keyout "$key_path" \
        -out "$cert_path" \
        -days 365 \
        -subj "/C=US/ST=California/L=San Francisco/O=Selki/CN=verify.selki.io"
}

# Function to set SSL file permissions
set_ssl_permissions() {
    chown "$GOPHISH_USER:$GOPHISH_USER" "$GOPHISH_HOME/ssl/"*
    chmod 600 "$GOPHISH_HOME/ssl/"*.key
}

# Function to setup SSL certificates
setup_ssl_certificates() {
    if [ ! -f "$GOPHISH_HOME/ssl/gophish.crt" ] || [ ! -f "$GOPHISH_HOME/ssl/gophish.key" ]; then
        log "Generating SSL certificates..."
        generate_ssl_certificate "gophish"
        generate_ssl_certificate "phish"
        set_ssl_permissions
    else
        log "SSL certificates already exist"
    fi
}

# Function to check service status
check_service_status() {
    sleep 2
    if systemctl is-active --quiet gophish.service; then
        log "Gophish service is running"
    else
        log "Error: Gophish service failed to start"
        systemctl status gophish.service --no-pager | tee -a "$LOG_FILE"
    fi
}

log "Starting Gophish setup..."

# Update package list
log "Updating package list..."
apt-get update -qq

# Install required packages
log "Installing required packages..."
apt-get install -y -qq wget curl git openssl libcap2-bin

# Install Go if not present or wrong version
if ! check_go_version "$GO_VERSION"; then
    log "Installing Go $GO_VERSION..."
    wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -O /tmp/go.tar.gz
    rm -rf /usr/local/go
    tar -C /usr/local -xzf /tmp/go.tar.gz
    rm /tmp/go.tar.gz
    
    # Add Go to PATH for all users
    if ! grep -q '/usr/local/go/bin' /etc/profile; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    fi
    export PATH=$PATH:/usr/local/go/bin
else
    log "Go $GO_VERSION or newer already installed"
fi

# Create gophish user if it doesn't exist
if ! id "$GOPHISH_USER" &>/dev/null; then
    log "Creating gophish user..."
    useradd -r -s /bin/false -d "$GOPHISH_HOME" -m "$GOPHISH_USER"
else
    log "User $GOPHISH_USER already exists"
fi

# Create gophish directory structure
log "Setting up directory structure..."
mkdir -p "$GOPHISH_HOME"/{ssl,static,templates,db}
chown -R "$GOPHISH_USER:$GOPHISH_USER" "$GOPHISH_HOME"

# Generate SSL certificates if they don't exist
setup_ssl_certificates

# Create config.json if it doesn't exist
if [ ! -f "$GOPHISH_HOME/config.json" ]; then
    log "Creating config.json..."
    cat > "$GOPHISH_HOME/config.json" <<EOF
{
    "admin_server": {
        "listen_url": "0.0.0.0:3333",
        "use_tls": true,
        "cert_path": "ssl/gophish.crt",
        "key_path": "ssl/gophish.key"
    },
    "phish_server": {
        "listen_url": "0.0.0.0:443",
        "use_tls": true,
        "cert_path": "ssl/phish.crt",
        "key_path": "ssl/phish.key"
    },
    "db_name": "sqlite3",
    "db_path": "db/gophish.db",
    "migrations_prefix": "db/db_sqlite3/",
    "contact_address": "",
    "logging": {
        "filename": "/var/log/gophish/gophish.log",
        "level": "info"
    }
}
EOF
    chown "$GOPHISH_USER:$GOPHISH_USER" "$GOPHISH_HOME/config.json"
else
    log "config.json already exists"
fi

# Create log directory
mkdir -p /var/log/gophish
chown "$GOPHISH_USER:$GOPHISH_USER" /var/log/gophish

# Allow gophish binary to bind to privileged ports
if [ -f "$GOPHISH_BINARY_PATH" ]; then
    log "Setting capabilities on gophish binary..."
    setcap 'cap_net_bind_service=+ep' "$GOPHISH_BINARY_PATH"
else
    log "Warning: Gophish binary not found at $GOPHISH_BINARY_PATH"
fi

# Create systemd service file
log "Creating systemd service..."
cat > /etc/systemd/system/gophish.service <<EOF
[Unit]
Description=Gophish Phishing Framework
After=network.target

[Service]
Type=simple
User=$GOPHISH_USER
Group=$GOPHISH_USER
WorkingDirectory=$GOPHISH_HOME
ExecStart=$GOPHISH_BINARY_PATH
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=gophish

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$GOPHISH_HOME /var/log/gophish
AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd daemon
systemctl daemon-reload

# Enable gophish service
log "Enabling gophish service..."
systemctl enable gophish.service

# Start or restart gophish service if binary exists
if [ -f "$GOPHISH_BINARY_PATH" ]; then
    log "Starting gophish service..."
    systemctl restart gophish.service
    check_service_status
else
    log "Gophish binary not found. Service will start once binary is deployed to $GOPHISH_BINARY_PATH"
fi

# Configure firewall if ufw is installed
if command_exists ufw; then
    log "Configuring firewall..."
    ufw allow 80/tcp
    ufw allow 443/tcp
    ufw allow 3333/tcp
fi

# Set up log rotation
log "Setting up log rotation..."
cat > /etc/logrotate.d/gophish <<EOF
/var/log/gophish/*.log {
    daily
    rotate 14
    compress
    delaycompress
    missingok
    notifempty
    create 0640 $GOPHISH_USER $GOPHISH_USER
    sharedscripts
    postrotate
        systemctl reload gophish.service > /dev/null 2>&1 || true
    endscript
}
EOF

# Create certificate renewal script
log "Creating certificate renewal script..."
cat > /opt/gophish/renew-certificates.sh <<EOF
#!/bin/bash
# Certificate renewal script for Gophish
set -euo pipefail

GOPHISH_HOME="$GOPHISH_HOME"
GOPHISH_USER="gophish"
LOG_FILE="/var/log/gophish-cert-renewal.log"

log() {
    echo "\$(date '+%Y-%m-%d %H:%M:%S') \$1" >> "\$LOG_FILE"
}

# Function to generate SSL certificates (shared with main script)
generate_ssl_certificate() {
    local cert_name="\$1"
    local key_path="\$GOPHISH_HOME/ssl/\${cert_name}.key"
    local cert_path="\$GOPHISH_HOME/ssl/\${cert_name}.crt"
    
    openssl req -x509 -newkey rsa:4096 -nodes \\
        -keyout "\$key_path" \\
        -out "\$cert_path" \\
        -days 365 \\
        -subj "/C=US/ST=California/L=San Francisco/O=Selki/CN=verify.selki.io"
}

# Function to set SSL file permissions (shared with main script)
set_ssl_permissions() {
    chown "\$GOPHISH_USER:\$GOPHISH_USER" "\$GOPHISH_HOME/ssl/"*
    chmod 600 "\$GOPHISH_HOME/ssl/"*.key
}

# Check if certificates are expiring within 30 days
check_cert_expiry() {
    local cert_file="\$1"
    if [ -f "\$cert_file" ]; then
        local expiry_date=\$(openssl x509 -in "\$cert_file" -noout -dates | grep notAfter | cut -d= -f2)
        local expiry_epoch=\$(date -d "\$expiry_date" +%s)
        local current_epoch=\$(date +%s)
        local days_until_expiry=\$(( (expiry_epoch - current_epoch) / 86400 ))
        
        if [ \$days_until_expiry -le 30 ]; then
            return 0  # Certificate expires within 30 days
        fi
    fi
    return 1  # Certificate is valid for more than 30 days
}

# Renew certificates
renew_certificates() {
    log "Renewing SSL certificates..."
    
    # Backup existing certificates
    if [ -f "\$GOPHISH_HOME/ssl/gophish.crt" ]; then
        cp "\$GOPHISH_HOME/ssl/gophish.crt" "\$GOPHISH_HOME/ssl/gophish.crt.bak"
        cp "\$GOPHISH_HOME/ssl/gophish.key" "\$GOPHISH_HOME/ssl/gophish.key.bak"
    fi
    if [ -f "\$GOPHISH_HOME/ssl/phish.crt" ]; then
        cp "\$GOPHISH_HOME/ssl/phish.crt" "\$GOPHISH_HOME/ssl/phish.crt.bak"
        cp "\$GOPHISH_HOME/ssl/phish.key" "\$GOPHISH_HOME/ssl/phish.key.bak"
    fi
    
    # Generate new certificates using shared functions
    generate_ssl_certificate "gophish"
    generate_ssl_certificate "phish"
    set_ssl_permissions
    
    # Restart gophish service
    systemctl restart gophish.service
    
    log "Certificate renewal completed successfully"
}

# Main execution
log "Starting certificate renewal check..."

if check_cert_expiry "\$GOPHISH_HOME/ssl/gophish.crt" || check_cert_expiry "\$GOPHISH_HOME/ssl/phish.crt"; then
    renew_certificates
else
    log "Certificates are still valid (more than 30 days remaining)"
fi
EOF

# Make renewal script executable
chmod +x /opt/gophish/renew-certificates.sh
chown root:root /opt/gophish/renew-certificates.sh

# Add cronjob for certificate renewal (runs weekly on Sunday at 2 AM)
log "Setting up certificate renewal cronjob..."
cat > /etc/cron.d/gophish-cert-renewal <<EOF
# Gophish certificate renewal - runs weekly on Sunday at 2 AM
0 2 * * 0 root /opt/gophish/renew-certificates.sh >/dev/null 2>&1
EOF

log "Gophish setup completed successfully!"
log "Note: Deploy the gophish binary to $GOPHISH_BINARY_PATH and run 'systemctl start gophish.service'"
log "Certificate renewal cronjob has been configured to run weekly on Sundays at 2 AM"