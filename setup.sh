echo -e "Macauth initialization..."

# 1. Directories
mkdir -p storage && mkdir -p storage/keys && mkdir -p storage/logs

# 2. Config file
if [ ! -f config.yaml ]; then
    echo -e "Waiting for config file..."
    if [ -f config.example.yaml ]; then
        cp config.example.yaml config.yaml
    else
        echo -e "There is no example file. Check GitHub repositrory: https://github.com/dmi3midd/macauth"
    fi
fi

# 3. RSA keys
if [ ! -f storage/keys/private.pem ] || [ ! -f storage/keys/public.pem ]; then
    echo -e "Waiting for RSA keys..."
    openssl genpkey -algorithm RSA -out storage/keys/private.pem -pkeyopt rsa_keygen_bits:2048 2>/dev/null
    openssl rsa -pubout -in storage/keys/private.pem -out storage/keys/public.pem 2>/dev/null
fi

# 4. Log files
if [ ! -f storage/logs/macauth.log ]; then
    echo -e "Waiting for log files..."
    touch storage/logs/macauth.log
fi

echo -e "Initialization is completed."
