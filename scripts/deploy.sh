#!/bin/bash
# Prior to starting, place init.sql and challengeFiles in home folder

DOMAIN="cannan.dev"
GH_USERNAME="moxlu"
SSH_PORT="8337"

sudo apt update
sudo apt upgrade -y
sudo apt install -y git wget certbot fail2ban opendkim opendkim-tools sqlite3

# Copy my public keys to the server
mkdir -p ~/.ssh
chmod 700 ~/.ssh
wget -qO ~/.ssh/authorized_keys https://github.com/$GH_USERNAME.keys 
chmod 600 ~/.ssh/authorized_keys

# Configure sshd
sudo cp /etc/ssh/sshd_config /etc/ssh/sshd_config.backup
sudo tee -a /etc/ssh/sshd_config.d/99-hardening.conf > /dev/null <<EOF
# SSH Security Configuration
Protocol 2
Port $SSH_PORT
StrictModes yes
PermitRootLogin no
PasswordAuthentication no
PubkeyAuthentication yes
AuthorizedKeysFile .ssh/authorized_keys
PermitEmptyPasswords no
ChallengeResponseAuthentication no
UsePAM yes
X11Forwarding no
PrintMotd no
ClientAliveInterval 300
ClientAliveCountMax 2
MaxAuthTries 3
MaxStartups 2
LoginGraceTime 60
AllowUsers $(whoami)
EOF

sudo systemctl reload sshd

# Configure ufw
sudo ufw --force reset
sudo ufw default deny incoming
sudo ufw default allow outgoing
# sudo ufw allow 25/tcp       # SMTP for postfix, only need this if accepting inbound
sudo ufw allow 80/tcp         # HTTP for certbot
sudo ufw allow 443/tcp        # HTTPS for cannan.dev
sudo ufw allow $SSH_PORT/tcp  # ssh
sudo ufw --force enable
sudo ufw status verbose

# Setup fail2ban for SSH protection
sudo systemctl enable fail2ban
sudo systemctl start fail2ban
sudo tee /etc/fail2ban/jail.local > /dev/null <<EOF
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 3

[sshd]
enabled = true
port = $SSH_PORT
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600
EOF

sudo systemctl restart fail2ban

# Initialise certbot for HTTPS
sudo certbot certonly --standalone -d $DOMAIN --non-interactive --agree-tos -m certbot@$DOMAIN
sudo chgrp -R ssl-cert /etc/letsencrypt/
sudo chmod -R 750 /etc/letsencrypt/
sudo find /etc/letsencrypt/ -type f -exec chmod 640 {} \;

# Install postfix for SMTP
sudo debconf-set-selections <<< "postfix postfix/mailname string mail.$DOMAIN"
sudo debconf-set-selections <<< "postfix postfix/main_mailer_type string 'Internet Site'"
sudo DEBIAN_FRONTEND=noninteractive apt install -y postfix
if systemctl is-active --quiet postfix; then
echo "Postfix installed and running"
sudo postconf mail_version
sudo postconf -e "inet_interfaces = loopback-only"
sudo postconf -e "inet_protocols = ipv4"
sudo postconf -e "mydestination = "
sudo postconf -e "local_recipient_maps = "
sudo postconf -e "local_transport = error:local mail delivery is disabled"
sudo postconf -e "mynetworks = 127.0.0.0/8"
sudo postconf -e "relay_domains = "
sudo postconf -e "default_transport = smtp"
sudo postconf -e "relay_transport = smtp"
sudo postconf -e "canonical_maps = hash:/etc/postfix/canonical"
sudo postconf -e "sender_canonical_maps = hash:/etc/postfix/sender_canonical"

# Set up canonical mapping to ensure emails come from noreply@cannan.dev
sudo tee /etc/postfix/canonical > /dev/null <<EOF
root@$(hostname) noreply@$DOMAIN
cannanbot@$(hostname) noreply@$DOMAIN
$(whoami)@$(hostname) noreply@$DOMAIN
EOF

sudo tee /etc/postfix/sender_canonical > /dev/null <<EOF
root noreply@$DOMAIN
cannanbot noreply@$DOMAIN
$(whoami) noreply@$DOMAIN
EOF

# Build the postfix hash tables
sudo postmap /etc/postfix/canonical
sudo postmap /etc/postfix/sender_canonical

sudo systemctl restart postfix
sudo systemctl enable postfix
echo "Postfix configured for send-only operation"
else
echo "Postfix installation failed"
fi

# Generate DKIM keys
sudo mkdir -p /etc/opendkim/keys/$DOMAIN
sudo opendkim-genkey -t -s default -d $DOMAIN -D /etc/opendkim/keys/$DOMAIN/

# Get cannan.dev from github
cd /opt/
if [ ! -d "cannan.dev" ]; then
    sudo git clone https://github.com/moxlu/cannan.dev.git
    sudo mkdir -p /opt/cannan.dev/run
fi

cd /opt/cannan.dev/run
sudo wget -qO /opt/cannan.dev/run/cannanapp https://github.com/moxlu/cannan.dev/releases/download/cannanapp/cannanapp
sudo chmod +x /opt/cannan.dev/run/cannanapp
sudo setcap 'cap_net_bind_service=+ep' /opt/cannan.dev/run/cannanapp # allows priv ports
sudo ln -sf /etc/letsencrypt/live/$DOMAIN/fullchain.pem ./fullchain.pem
sudo ln -sf /etc/letsencrypt/live/$DOMAIN/privkey.pem ./privkey.pem
openssl rand -base64 32 > session.key

if [ ! -f cannan.db ]; then
    sudo sqlite3 cannan.db < /opt/cannan.dev/scripts/schema.sql
fi

cd ~
if [ -d "./challengeFiles" ]; then
    sudo mv ./challengeFiles /opt/cannan.dev
    echo "challengeFiles moved successfully"
else
    echo "Missing challengeFiles in home directory"
fi

if [ -f "./init.sql" ]; then
    sudo mv ./init.sql /opt/cannan.dev/run/
    echo "init.sql moved successfully"
    cd /opt/cannan.dev/run/
    sudo sqlite3 cannan.db < init.sql
else
    echo "Missing init.sql in home directory"
fi

sudo cp /opt/cannan.dev/scripts/cannan.service /etc/systemd/system/cannan.service
sudo cp /opt/cannan.dev/scripts/restart_cannan.sh /etc/letsencrypt/renewal-hooks/deploy/restart_cannan.sh
sudo chmod +x /etc/letsencrypt/renewal-hooks/deploy/restart_cannan.sh

id -u cannanbot &>/dev/null || sudo useradd -r -s /bin/false cannanbot
sudo groupadd -f ssl-cert
sudo usermod -aG ssl-cert cannanbot
sudo chown -R cannanbot:cannanbot /opt/cannan.dev

sudo systemctl daemon-reload
sudo systemctl enable cannan.service
sudo systemctl start cannan.service

echo "🎉 Deployment complete! Current service status:"
sudo systemctl status cannan.service --no-pager -l
echo ""
echo "Rebooting in 10 seconds to test service startup."
echo "Connect via: ssh -p $SSH_PORT $(whoami)@$DOMAIN"
sleep 10
sudo reboot