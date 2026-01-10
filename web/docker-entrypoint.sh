#!/bin/sh

SSL_DIR="/etc/nginx/ssl"
CERT_FILE="$SSL_DIR/server.crt"
KEY_FILE="$SSL_DIR/server.key"

# 创建 SSL 目录
mkdir -p $SSL_DIR

# 如果证书不存在，生成自签证书
if [ ! -f "$CERT_FILE" ] || [ ! -f "$KEY_FILE" ]; then
    echo "Generating self-signed SSL certificate..."
    openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
        -keyout $KEY_FILE \
        -out $CERT_FILE \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=CSCAN/OU=IT/CN=localhost" \
        -addext "subjectAltName=DNS:localhost,DNS:*.localhost,IP:127.0.0.1"
    echo "SSL certificate generated successfully."
else
    echo "SSL certificate already exists, skipping generation."
fi

# 启动 nginx
exec nginx -g "daemon off;"
