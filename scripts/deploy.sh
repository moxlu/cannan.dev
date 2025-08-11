#!/bin/bash

sudo apt update
sudo apt upgrade -y
sudo apt install -y sqlite3 certbot git wget golang-go

# setup sshd
# setup ufw

sudo certbot certonly --standalone -d cannan.dev --non-interactive --agree-tos -m certs@cannan.dev
sudo chgrp -R ssl-cert /etc/letsencrypt/
sudo chmod -R 750 /etc/letsencrypt/
sudo find /etc/letsencrypt/ -type f -exec chmod 640 {} \;

cd /opt/
if [ ! -d "cannan.dev" ]; then
    sudo git clone https://github.com/moxlu/cannan.dev.git
fi

sudo mkdir -p /opt/cannan.dev/run
cd /opt/cannan.dev/run

sudo wget -O /opt/cannan.dev/run/cannanapp https://github.com/moxlu/cannan.dev/releases/download/cannanapp/cannanapp
sudo chmod +x /opt/cannan.dev/cannanapp
sudo setcap 'cap_net_bind_service=+ep' /opt/cannan.dev/run/cannanapp # allows priv ports

sudo ln -s /etc/letsencrypt/live/cannan.dev/fullchain.pem ./fullchain.pem
sudo ln -s /etc/letsencrypt/live/cannan.dev/privkey.pem ./privkey.pem
sudo sqlite3 cannan.db < ../scripts/schema.sql

# from dev: scp and mv init.sql and challengeFiles

cd /opt/cannan.dev/run
sudo sqlite3 cannan.db < init.sql

cd /opt/cannan.dev/scripts
sudo ./generate_session_key.sh
sudo cp cannan.service /etc/systemd/system/cannan.service
sudo cp restart_cannan.sh /etc/letsencrypt/renewal-hooks/deploy/restart_cannan.sh
sudo chmod +x /etc/letsencrypt/renewal-hooks/deploy/restart_cannan.sh

sudo useradd -r -s /bin/false cannanbot
sudo groupadd ssl-cert
sudo usermod -aG ssl-cert cannanbot
sudo chown -R cannanbot:cannanbot /opt/cannan.dev

sudo systemctl daemon-reload
sudo systemctl enable cannan.service
sudo systemctl start cannan.service