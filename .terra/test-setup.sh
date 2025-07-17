#!/bin/bash
set -e

echo "Testing Gophish setup script..."

# Run in Docker container without systemd
docker run --rm -v $(pwd)/setup-gophish.sh:/setup-gophish.sh:ro ubuntu:22.04 bash -c "
    # Modify the script to skip systemd-specific operations
    sed 's/systemctl/echo \"Would run: systemctl\"/g' /setup-gophish.sh > /tmp/setup-test.sh
    chmod +x /tmp/setup-test.sh
    
    # Run the modified script
    /tmp/setup-test.sh
    
    # Verify installations
    echo '=== Verification ==='
    echo -n 'Go version: '
    /usr/local/go/bin/go version || echo 'Go not installed'
    
    echo -n 'User gophish exists: '
    id gophish >/dev/null 2>&1 && echo 'YES' || echo 'NO'
    
    echo -n 'Directory structure: '
    ls -la /opt/gophish/ 2>/dev/null || echo 'Not created'
    
    echo -n 'SSL certificates: '
    ls -la /opt/gophish/ssl/*.crt 2>/dev/null | wc -l | xargs -I {} echo '{} certificates found'
    
    echo -n 'Config file exists: '
    [ -f /opt/gophish/config.json ] && echo 'YES' || echo 'NO'
    
    echo -n 'Log directory exists: '
    [ -d /var/log/gophish ] && echo 'YES' || echo 'NO'
"