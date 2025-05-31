#!/bin/bash

# Immediately exit the script if any command fails 
set -e

# Ensuring /usr/sbin is in PATH to find admin binaries like nginx
export PATH=$PATH:/usr/sbin

pwd

EMAILID=thisisvoiddd1@gmail.com
DOMAIN="brainwars.thisisvoid.in" 
SERVICE_NAME="brainwars.service"
LOCAL_SERVICE_PATH="./deployment/brainwars.service"
SYSTEMD_PATH="/etc/systemd/system/$SERVICE_NAME"

# moving the prod config to config
cp ./config/configlist/config-prod.json ./config/config.json

# Install Go if not present
if ! command -v go &> /dev/null; then
    echo "Go not found. Installing..."
    curl -OL https://go.dev/dl/go1.24.3.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf go1.24.3.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
    source ~/.profile
    rm go1.24.3.linux-amd64.tar.gz
else
    echo "Go is already installed: $(go version)"
fi


# Initialize Go module if not already initialized
if [ ! -f go.mod ]; then
    echo "Initializing Go module 'brainwars'..."
    go mod init brainwars
else
    echo "go.mod already exists. Skipping 'go mod init'."
fi

# Tidy dependencies
go mod tidy

# Install goose if not present
if ! command -v goose >/dev/null 2>&1; then
    echo "Installing goose..."
    go install github.com/pressly/goose/v3/cmd/goose@latest
   echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.profile
   source ~/.profile
else
    echo "goose already installed."
fi

# Install sqlc if not present
if ! command -v sqlc >/dev/null 2>&1; then
    echo "Installing sqlc..."
   sudo snap install sqlc 
#    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
else
    echo "sqlc already installed."
fi

# install tailwind css and minify for production
if [ ! -f tailwindcss ]; then 
    echo "installing tailwind css.........."
    curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64
    mv tailwindcss-linux-x64 tailwindcss 
    chmod +x tailwindcss
else 
    echo "tailwind css already exists"
fi
./tailwindcss -i web/ui/utility/css/input.css -o web/ui/utility/css/output.css --config web/ui/tailwind.config.js --minify 
# ./tailwindcss init
# rm tailwindcss
echo "successful"

# Install nginx if not present
if ! command -v nginx >/dev/null 2>&1; then
    echo "Installing nginx..."
    sudo apt-get clean
    sudo rm -rf /var/lib/apt/lists/*
    sudo apt-get update
    sudo apt-get install -y nginx
    export PATH=$PATH:/usr/sbin
else
    echo "nginx already installed."
fi

# Install docker if not present
if ! command -v docker >/dev/null 2>&1; then
    echo "Docker not found. Installing Docker..."

    # Update package list and install prerequisites
    sudo apt-get update
    sudo apt-get install -y \
        ca-certificates \
        curl \
        gnupg \
        lsb-release
    for pkg in docker.io docker-doc docker-compose docker-compose-v2 podman-docker containerd runc; do sudo apt-get remove $pkg; done
    # Add Docker's official GPG key:
    sudo apt-get update
    sudo apt-get install ca-certificates curl
    sudo install -m 0755 -d /etc/apt/keyrings
    sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
    sudo chmod a+r /etc/apt/keyrings/docker.asc

    # Add the repository to Apt sources:
    echo \
    "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
    $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}") stable" | \
    sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
    sudo apt-get update

    sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
    sudo docker run hello-world

    echo "Docker installed successfully."
else
    echo "Docker already installed."
fi


#check if prod environment variable file exists
if [ -f .env.production ];then
    echo "Production environment variables available. so using them"
    source .env.production
else
    echo "Production environment variables not available. exiting"
exit 1;
fi
# run docker compose to start the postgres database
sudo docker compose -f ./deployment/docker/docker-compose-pgsql.yml up -d


# run goose to apply migrations
echo "Applying goose database migrations on our postgres container..."
goose -dir migrations postgres "host=${PG_HOST} port=${PG_PORT} user=${PG_USER} password=${PG_PASSWORD} dbname=${PG_DB} sslmode=${PG_SSLMODE}" up

# setting up nginx and certbot for certficate
#installing certbot
if ! command -v certbot &> /dev/null; then
    sudo snap install --classic certbot
    sudo ln -s /snap/bin/certbot /usr/bin/certbot
    echo "Certbot installed...."
else
    echo "Certbot already installed..."
fi

echo "Getting Let's Encrypt certificate..."
echo "Make sure $DOMAIN points to this server's public IP!"


# Create temporary basic config for certbot
sudo tee /etc/nginx/nginx.conf > /dev/null << EOF
worker_processes auto;
events { worker_connections 1024; }
http {
    include mime.types;
    server {
        listen 80;
        server_name $DOMAIN;
        location / {
            return 200 "OK";
            add_header Content-Type text/plain;
        }
    }
}
EOF

# Test and reload nginx
sudo nginx -t
if pgrep nginx > /dev/null; then
    sudo nginx -s reload
else
    sudo nginx
fi

# Get certificate (skip if already exists)
if [[ ! -d "/etc/letsencrypt/live/$DOMAIN" ]]; then
    echo "Getting new certificate..."    
    if [[ -z "$EMAILID" ]]; then
        echo "Email is required for production use!"
        exit 1
    fi
    
    sudo certbot --nginx -d "$DOMAIN" --non-interactive --agree-tos --email "$EMAILID"
else
    echo "Certificate already exists..."
fi

echo "Testing nginx configuration syntax..."
sudo nginx -t

echo "Setting permissions for SSL certs..."
sudo chmod 755 /etc/letsencrypt/{live,archive}
sudo chmod 644 /etc/letsencrypt/live/$DOMAIN/*.pem

sudo cp ./deployment/nginx.conf /etc/nginx/nginx.conf

if pgrep nginx > /dev/null; then
    echo "Nginx is running. Reloading..."
    sudo nginx -s reload
else
    echo "Starting nginx..."
    sudo nginx
fi

echo "Stated nginx successfully"

# start the backend server
echo "Starting the backend go server through a systemd timer..."
# Check if local service file exists
if [ ! -f "$LOCAL_SERVICE_PATH" ]; then
  echo "ERROR: $LOCAL_SERVICE_PATH does not exist."
  exit 1
fi

# Copy service file to systemd directory
echo "Copying $SERVICE_NAME to $SYSTEMD_PATH"
sudo cp "$LOCAL_SERVICE_PATH" "$SYSTEMD_PATH"

# Reload systemd to recognize changes
echo "Reloading systemd daemon..."
sudo systemctl daemon-reload

# Enable service to start on boot
echo "Enabling $SERVICE_NAME..."
sudo systemctl enable "$SERVICE_NAME"

# Check current status
STATUS=$(systemctl is-active "$SERVICE_NAME" || echo "inactive")

if [ "$STATUS" == "active" ]; then
  echo  "Service is already running. Restarting..."
  sudo systemctl restart "$SERVICE_NAME"
else
  echo "Service is not running. Starting..."
  sudo systemctl start "$SERVICE_NAME"
fi

# Show final status
echo "Current status:"
systemctl status "$SERVICE_NAME" --no-pager