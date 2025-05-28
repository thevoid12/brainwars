#!/bin/bash

# Immediately exit the script if any command fails 
set -e

# Ensuring /usr/sbin is in PATH to find admin binaries like nginx
export PATH=$PATH:/usr/sbin

pwd

# moving the prod config to config
cp ./config/configlist/config-prod.json ./config/config.json

# Install Go if not present
if ! command -v go >/dev/null 2>&1; then
    echo "Go not found. Installing..."
    curl -OL https://go.dev/dl/go1.24.3.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf go1.24.3.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    source ~/.bashrc
else
    echo "Go is already installed."
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
else
    echo "goose already installed."
fi

# Install sqlc if not present
if ! command -v sqlc >/dev/null 2>&1; then
    echo "Installing sqlc..."
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
else
    echo "sqlc already installed."
fi

# install tailwind css and minify for production
if ! command -v tailwind >/dev/null 2>&1; then 
    echo "installing tailwind css.........."
else 
    echo "tailwind css already exists"
fi
curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64
mv tailwindcss-linux-x64 tailwindcss 
chmod +x tailwindcss
./tailwindcss -i web/ui/utility/css/input.css -o web/ui/utility/css/output.css --config web/ui/tailwind.config.js --minify 
# ./tailwindcss init
# rm tailwindcss
echo "successful"
exit 1

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
docker compose -f ./deployment/docker/docker-compose-pgsql.yml up -d


# run goose to apply migrations
echo "Applying goose database migrations on our postgres container..."
goose -dir migrations postgres "host=${PG_HOST} port=${PG_PORT} user=${PG_USER} password=${PG_PASSWORD} dbname=${PG_DB} sslmode=${PG_SSLMODE}" up

# setting up nginx
echo "Setting up nginx..."
sudo cp ./deployment/nginx.conf /etc/nginx/nginx.conf
if pidof nginx >/dev/null; then
    echo "Nginx is running. Restarting..."
    sudo nginx -s reload
else
    echo "Starting Nginx.."
    sudo nginx
fi

# Find the PID of any existing brainwars process
PID=$(pgrep -f "./brainwars")

if [ -n "$PID" ]; then
    echo "brainwars is already running with PID $PID. Stopping..."
    kill "$PID"
    sleep 1  # give it a moment to shut down
fi

# start the backend server
echo "Starting the backend go server..."
go build brainwars
# run in background
./brainwars    