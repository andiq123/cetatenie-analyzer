#!/bin/bash

# Exit on error
set -e

echo "🚀 Starting deployment process..."

# Set environment variables for ARM64 optimization
export GOARCH=arm64
export GOOS=linux
export CGO_ENABLED=1

# Install required system dependencies
echo "📦 Installing system dependencies..."
sudo apt-get update
sudo apt-get install -y \
    gcc-aarch64-linux-gnu \
    libsqlite3-dev \
    build-essential

# Clean previous build
echo "🧹 Cleaning previous build..."
rm -rf build/
mkdir -p build

# Build the application with optimizations
echo "🔨 Building application..."
go build -tags sqlite3 \
    -ldflags="-s -w" \
    -o build/cetatenie-analyzer \
    ./cmd/main.go

# Create systemd service file
echo "📝 Creating systemd service..."
sudo tee /etc/systemd/system/cetatenie-analyzer.service << EOF
[Unit]
Description=Cetatenie Analyzer Bot
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=/home/$USER/cetatenie-analyzer
ExecStart=/home/$USER/cetatenie-analyzer/build/cetatenie-analyzer
Restart=always
RestartSec=10
Environment=TELEGRAM_BOT_TOKEN=your_bot_token_here

[Install]
WantedBy=multi-user.target
EOF

# Set proper permissions
echo "🔒 Setting permissions..."
chmod +x build/cetatenie-analyzer

# Reload systemd and restart service
echo "🔄 Reloading systemd and restarting service..."
sudo systemctl daemon-reload
sudo systemctl restart cetatenie-analyzer
sudo systemctl enable cetatenie-analyzer

echo "✅ Deployment completed!"
echo "📊 Service status:"
sudo systemctl status cetatenie-analyzer

# Show logs
echo "📋 Recent logs:"
sudo journalctl -u cetatenie-analyzer -n 50 --no-pager 