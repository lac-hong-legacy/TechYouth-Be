#!/bin/bash
set -e

echo "🚀 Setting up VEN API staging server..."
echo ""

# Update system
echo "📦 Updating system packages..."
sudo apt-get update -y
sudo apt-get upgrade -y

# Install Docker
if ! command -v docker &> /dev/null; then
    echo "🐳 Installing Docker..."
    curl -fsSL https://get.docker.com | sh
    sudo usermod -aG docker $USER
    echo "✅ Docker installed"
else
    echo "✅ Docker already installed ($(docker --version))"
fi

# Install Docker Compose
if ! docker compose version &> /dev/null; then
    echo "🐳 Installing Docker Compose..."
    sudo apt-get install -y docker-compose-plugin
    echo "✅ Docker Compose installed"
else
    echo "✅ Docker Compose already installed ($(docker compose version))"
fi

# Install Git
if ! command -v git &> /dev/null; then
    echo "📚 Installing Git..."
    sudo apt-get install -y git
fi

# Create app directory
echo "📁 Creating application directory..."
sudo mkdir -p /opt/ven-api
sudo chown $USER:$USER /opt/ven-api

# Clone repository
cd /opt/ven-api
if [ ! -d .git ]; then
    echo "📥 Cloning repository..."
    git clone https://github.com/lac-hong-legacy/TechYouth-Be.git .
else
    echo "✅ Repository already cloned"
    git pull origin main
fi

# Setup .env
if [ ! -f .env ]; then
    echo "⚙️  Creating .env file..."
    cp .env.example .env
    echo ""
    echo "⚠️  IMPORTANT: Edit /opt/ven-api/.env with your configuration!"
    echo ""
else
    echo "✅ .env file exists"
fi

# Setup SSH directory
echo "🔑 Setting up SSH..."
mkdir -p ~/.ssh
chmod 700 ~/.ssh
touch ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys

# Setup firewall
if command -v ufw &> /dev/null; then
    echo "🔥 Configuring firewall..."
    sudo ufw --force enable
    sudo ufw allow 22/tcp   # SSH
    sudo ufw allow 80/tcp   # HTTP
    sudo ufw allow 443/tcp  # HTTPS
    sudo ufw allow 8000/tcp # API
    echo "✅ Firewall configured"
fi

# Create backup directory
mkdir -p /opt/ven-api/backups

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Setup complete!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📋 Next steps:"
echo ""
echo "1. Edit configuration:"
echo "   vim /opt/ven-api/.env"
echo ""
echo "2. Add your SSH public key:"
echo "   vim ~/.ssh/authorized_keys"
echo "   (Paste the public key from your local machine)"
echo ""
echo "3. Start the application:"
echo "   cd /opt/ven-api"
echo "   docker compose up -d"
echo ""
echo "4. Check status:"
echo "   docker compose ps"
echo "   docker compose logs -f ven-api"
echo ""
echo "5. Test health endpoint:"
echo "   curl http://localhost:8000/health"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
