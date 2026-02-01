#!/bin/bash

# setup-proxy.sh - Setup Tesla HTTP Proxy with Docker Compose

set -e

echo "🚀 Tesla HTTP Proxy Setup"
echo "========================="
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed"
    echo "📖 Please install Docker from https://www.docker.com/products/docker-desktop"
    exit 1
fi

echo "✅ Docker found"
echo ""

# Create config directory
mkdir -p config

# Check if TLS certificates exist
if [ ! -f config/tls-cert.pem ] || [ ! -f config/tls-key.pem ]; then
    echo "🔐 Generating TLS certificates..."
    
    # Create OpenSSL config file for compatibility with Windows
    cat > config/openssl.conf <<'OPENSSL_EOF'
[req]
default_bits = 2048
prompt = no
default_md = sha256
distinguished_name = req_distinguished_name
req_extensions = v3_req

[req_distinguished_name]
CN = localhost

[v3_req]
extendedKeyUsage = serverAuth
keyUsage = digitalSignature, keyCertSign, keyAgreement
OPENSSL_EOF

    openssl req -x509 -nodes -newkey ec \
        -pkeyopt ec_paramgen_curve:secp384r1 \
        -pkeyopt ec_param_enc:named_curve \
        -config config/openssl.conf \
        -keyout config/tls-key.pem \
        -out config/tls-cert.pem \
        -sha256 -days 3650
    
    rm config/openssl.conf
    echo "✅ TLS certificates generated"
else
    echo "✅ TLS certificates already exist"
fi

echo ""

# Check if private key exists
if [ ! -f config/fleet-key.pem ]; then
    echo "❌ fleet-key.pem not found in config/ directory"
    echo "📝 Please copy your private key:"
    echo "   cp private-key.pem config/fleet-key.pem"
    exit 1
fi

echo "✅ Fleet private key found"
echo ""

# Check if tokens file exists
if [ ! -f tesla-tokens.json ]; then
    echo "⚠️  tesla-tokens.json not found. It will be created when you authenticate."
    echo ""
fi

echo "📋 Configuration Summary:"
echo "  - TLS Certificates: config/tls-cert.pem, config/tls-key.pem"
echo "  - Fleet Private Key: config/fleet-key.pem"
echo "  - Proxy Port: 4443"
echo "  - Tokens File: tesla-tokens.json"
echo ""

echo "🐳 Pulling Tesla vehicle-command Docker image..."
docker pull tesla/vehicle-command:latest

echo ""
echo "✅ Setup complete!"
echo ""
echo "To start the proxy, run:"
echo "  docker-compose up -d"
echo ""
echo "To view logs:"
echo "  docker-compose logs -f tesla-proxy"
echo ""
echo "To stop the proxy:"
echo "  docker-compose down"
