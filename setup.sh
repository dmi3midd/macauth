echo -e "Macauth initialization..."

# 1. Directories
mkdir -p keys storage

# 2. API key
# if [ ! -f .env ]; then
#     echo -e "Waiting for API_KEY..."
#     API_KEY=$(openssl rand -hex 32)
#     echo "API_KEY=$API_KEY" > .env
# fi

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
if [ ! -f keys/private.pem ] || [ ! -f keys/public.pem ]; then
    echo -e "Waiting for RSA keys..."
    openssl genpkey -algorithm RSA -out keys/private.pem -pkeyopt rsa_keygen_bits:2048 2>/dev/null
    openssl rsa -pubout -in keys/private.pem -out keys/public.pem 2>/dev/null
fi

# 5. Database file
if [ ! -f storage/db.sql ]; then
    echo -e "Waiting for database file..."
    touch storage/macauth.db
fi

echo -e "Initialization is completed."
