echo -e "Macauth initialization..."

# 1. Directories
mkdir -p storage && mkdir -p storage/keys

# 2. API key
if [ ! -f .env ] || ! grep -q "^API_KEY=" .env; then
    echo -e "Waiting for API_KEY..."
    API_KEY=$(openssl rand -hex 32)
    echo "API_KEY=$API_KEY" >> .env
fi

# 3. Config file
if [ ! -f config.yaml ]; then
    echo -e "Waiting for config file..."
    if [ -f config.example.yaml ]; then
        cp config.example.yaml config.yaml
    else
        echo -e "There is no example file. Check GitHub repositrory: https://github.com/dmi3midd/macauth"
    fi
fi

# 4. RSA keys
if [ ! -f storage/keys/private.pem ] || [ ! -f storage/keys/public.pem ]; then
    echo -e "Waiting for RSA keys..."
    openssl genpkey -algorithm RSA -out storage/keys/private.pem -pkeyopt rsa_keygen_bits:2048 2>/dev/null
    openssl rsa -pubout -in storage/keys/private.pem -out storage/keys/public.pem 2>/dev/null
fi

# 5. Database and log files
if [ ! -f storage/macauth.db ]; then
    echo -e "Waiting for database and log files..."
    touch storage/macauth.db
    touch storage/macauth.log
fi

echo -e "Initialization is completed."
