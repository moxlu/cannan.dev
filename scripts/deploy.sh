#!/bin/bash

# setup sshd
# setup ufw

sudo apt update
sudo apt upgrade -y
sudo apt install -y golang-go sqlite3 certbot git

if ! id -u cannanbot > /dev/null 2>&1; then
    sudo useradd -r -s /bin/false cannanbot
fi
sudo groupadd ssl-cert
sudo usermod -aG ssl-cert cannanbot

sudo certbot certonly --standalone -d cannan.dev --non-interactive --agree-tos -m certs@cannan.dev
sudo chgrp -R ssl-cert /etc/letsencrypt/live/cannan.dev
sudo chgrp -R ssl-cert /etc/letsencrypt/archive/cannan.dev
sudo chmod -R 750 /etc/letsencrypt/live/cannan.dev
sudo chmod -R 750 /etc/letsencrypt/archive/cannan.dev
sudo find /etc/letsencrypt/live/cannan.dev -type f -exec chmod 640 {} \;
sudo find /etc/letsencrypt/archive/cannan.dev -type f -exec chmod 640 {} \;

cd /opt/
if [ ! -d "cannan.dev" ]; then
    sudo git clone https://github.com/moxlu/cannan.dev.git
fi

cd /opt/cannan.dev
sudo -u cannanbot go build -o cannanapp ./src
sudo mkdir -p run challengeFiles
sudo chown -R cannanbot:cannanbot /opt/cannan.dev

cd /opt/cannan.dev/run
sudo ln -s /etc/letsencrypt/live/cannan.dev/fullchain.pem ./fullchain.pem
sudo ln -s /etc/letsencrypt/live/cannan.dev/privkey.pem ./privkey.pem
sudo sqlite3 cannan.db < ../scripts/schema.sql

cd /opt/cannan.dev/scripts
sudo ./generate_session_key.sh
sudo cp cannan.service /etc/systemd/system/cannan.service
sudo cp restart_cannan.sh /etc/letsencrypt/renewal-hooks/deploy/restart-cannan.sh
sudo chmod +x /etc/letsencrypt/renewal-hooks/deploy/restart_cannan.sh

sudo chown -R cannanbot:cannanbot /opt/cannan.dev

sudo systemctl daemon-reload
sudo systemctl enable cannan.service
sudo systemctl start cannan.service

# Don't forget to scp init.sql and challengeFiles