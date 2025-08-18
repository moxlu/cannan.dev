#!/bin/bash

# Prior to starting, place this script, init.sql and challengeFiles in /opt. From local:
# cd cannan.dev
# scp ./scripts/deploy.sh root@IP:/opt/
# scp ./run/init.sql root@IP:/opt/
# scp -r ./challengeFiles root@IP:/opt/
# ssh root@IP
# cd /opt && ls
# chmod +x ./deploy.sh
# ./deploy.sh

read -p "Domain name for this server: " DOMAIN
read -p "Select an SSH port for this server: " SSH_PORT
read -p "Enter an email address for certbot & postfix test: " TEST_EMAIL
read -p "Enter your github username: " GH_USERNAME
echo "Creating human user $GH_USERNAME"
id -u $GH_USERNAME &>/dev/null || sudo useradd -m -s /bin/bash $GH_USERNAME
usermod -aG sudo $GH_USERNAME
sudo passwd $GH_USERNAME

echo "Copying Github public keys to user's authorized_keys"
mkdir -p /home/$GH_USERNAME/.ssh
chmod 700 /home/$GH_USERNAME/.ssh
wget -qO /home/$GH_USERNAME/.ssh/authorized_keys https://github.com/$GH_USERNAME.keys 
chmod 600 /home/$GH_USERNAME/.ssh/authorized_keys
chown -R $GH_USERNAME:$GH_USERNAME /home/$GH_USERNAME/.ssh

echo "Creating bot user cannanbot"
id -u cannanbot &>/dev/null || sudo useradd -r -s /bin/false cannanbot
sudo groupadd -f ssl-cert
sudo usermod -aG ssl-cert cannanbot

sudo apt update && sudo apt upgrade -y
sudo DEBIAN_FRONTEND=noninteractive apt install -y git wget certbot fail2ban sqlite3 libcap2-bin

echo "Configuring sshd"
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
AllowUsers $GH_USERNAME
EOF

sudo systemctl reload ssh || sudo systemctl reload sshd

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
sudo certbot certonly --standalone -d $DOMAIN --non-interactive --agree-tos -m $TEST_EMAIL
sudo chgrp -R ssl-cert /etc/letsencrypt/
sudo chmod -R 750 /etc/letsencrypt/
sudo find /etc/letsencrypt/ -type f -exec chmod 640 {} \;

# OpenDKIM before postfix












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
$GH_USERNAME@$(hostname) noreply@$DOMAIN
EOF

sudo tee /etc/postfix/sender_canonical > /dev/null <<EOF
root noreply@$DOMAIN
cannanbot noreply@$DOMAIN
$GH_USERNAME noreply@$DOMAIN
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

echo "Configuring OpenDKIM for Postfix"
sudo DEBIAN_FRONTEND=noninteractive apt install -y opendkim opendkim-tools

# Create key directory and runtime directory
sudo mkdir -p /etc/opendkim/keys/$DOMAIN
sudo chown -R opendkim:opendkim /etc/opendkim
sudo mkdir -p /run/opendkim
sudo chown opendkim:opendkim /run/opendkim
sudo chmod 755 /run/opendkim

# Generate DKIM key (only if it doesn't exist)
if [ ! -f /etc/opendkim/keys/$DOMAIN/default.private ]; then
    sudo opendkim-genkey -t -s default -d $DOMAIN -D /etc/opendkim/keys/$DOMAIN/
    sudo chown opendkim:opendkim /etc/opendkim/keys/$DOMAIN/*
fi

# OpenDKIM configuration
sudo tee /etc/opendkim.conf > /dev/null <<EOF
Syslog          yes
UMask           002
Canonicalization relaxed/simple
Mode            sv
SubDomains      no
AutoRestart     yes
AutoRestartRate 10/1h
Background      yes
DNSTimeout      5
SignatureAlgorithm rsa-sha256
PIDFile         /run/opendkim/opendkim.pid
Socket          unix:/run/opendkim/opendkim.sock
KeyTable        /etc/opendkim/key.table
SigningTable    /etc/opendkim/signing.table
ExternalIgnoreList /etc/opendkim/trusted.hosts
InternalHosts   /etc/opendkim/trusted.hosts
EOF

# KeyTable
sudo tee /etc/opendkim/key.table > /dev/null <<EOF
default._domainkey.$DOMAIN $DOMAIN:default:/etc/opendkim/keys/$DOMAIN/default.private
EOF

# SigningTable
sudo tee /etc/opendkim/signing.table > /dev/null <<EOF
*@${DOMAIN} default._domainkey.$DOMAIN
EOF

# Trusted hosts
sudo tee /etc/opendkim/trusted.hosts > /dev/null <<EOF
127.0.0.1
localhost
$DOMAIN
EOF

# Add postfix user to opendkim group
sudo usermod -aG opendkim postfix

# Update Postfix to use OpenDKIM
sudo postconf -e "milter_default_action = accept"
sudo postconf -e "milter_protocol = 6"
sudo postconf -e "smtpd_milters = unix:/run/opendkim/opendkim.sock"
sudo postconf -e "non_smtpd_milters = \$smtpd_milters"

# Enable and start service
sudo systemctl enable opendkim
sudo systemctl restart opendkim
sudo systemctl restart postfix
echo "DKIM setup completed."

# Testmail
cat <<EOF | sudo tee /etc/profile.d/testmail.sh
alias testmail='sendmail $TEST_EMAIL <<< "Subject: Test $DOMAIN \$(date +%H:%M)
From: noreply@$DOMAIN

Test sent at \$(date).
Please check the headers to ensure authentication is working."'
EOF
sudo chmod +x /etc/profile.d/testmail.sh
testmail

# Get cannan.dev from github
cd /opt/
if [ ! -d "cannan.dev" ]; then
    sudo git clone https://github.com/moxlu/cannan.dev.git
    sudo mkdir -p /opt/cannan.dev/run
fi

if [ -d "./challengeFiles" ]; then
    sudo mv ./challengeFiles /opt/cannan.dev
    echo "challengeFiles moved successfully"
else
    echo "Missing challengeFiles in /opt directory"
fi

if [ -f "./init.sql" ]; then
    sudo mv ./init.sql /opt/cannan.dev/run/
    echo "init.sql moved successfully"
else
    echo "Missing init.sql in /opt directory"
fi

cd /opt/cannan.dev/run

# Download cannanapp
APP_PATH="/opt/cannan.dev/run/cannanapp"
LATEST_TAG=$(wget -qO- https://api.github.com/repos/moxlu/cannan.dev/releases/latest | jq -r .tag_name)
sudo wget -qO "$APP_PATH" https://github.com/moxlu/cannan.dev/releases/download/$LATEST_TAG/cannanapp
sudo chmod +x "$APP_PATH"
sudo sync # sometime setcap doesn't work and this might help
sudo setcap 'cap_net_bind_service=+ep' "$APP_PATH" # reqd for priv ports like :443

# Verify it worked
CAP_CHECK=$(getcap "$APP_PATH")
if [[ -z "$CAP_CHECK" ]]; then
    echo "Warning: setcap failed on $APP_PATH"
else
    echo "setcap applied successfully: $CAP_CHECK"
fi

# Keys
sudo ln -sf /etc/letsencrypt/live/$DOMAIN/fullchain.pem ./fullchain.pem
sudo ln -sf /etc/letsencrypt/live/$DOMAIN/privkey.pem ./privkey.pem
openssl rand -base64 32 > session.key

if [ ! -f cannan.db ]; then
    sudo sqlite3 cannan.db < /opt/cannan.dev/scripts/schema.sql
    sudo sqlite3 cannan.db < init.sql
fi

sudo cp /opt/cannan.dev/scripts/cannan.service /etc/systemd/system/cannan.service
sudo cp /opt/cannan.dev/scripts/restart_cannan.sh /etc/letsencrypt/renewal-hooks/deploy/restart_cannan.sh
sudo chmod +x /etc/letsencrypt/renewal-hooks/deploy/restart_cannan.sh

sudo chown -R cannanbot:cannanbot /opt/cannan.dev

sudo systemctl daemon-reload
sudo systemctl enable cannan.service
sudo systemctl start cannan.service

echo "Deployment complete! Current service status:"
sudo systemctl status cannan.service --no-pager -l
echo ""
echo "========================================================="
echo "Add the following DNS records for your domain:"
echo "MX 10 mail.$DOMAIN"
echo "_spf.$DOMAIN IN TXT \"v=spf1 ip4:$(curl -s ifconfig.me) -all\""
echo "_dmarc.$DOMAIN IN TXT \"v=DMARC1; p=quarantine; rua=mailto:dmarc@$DOMAIN\""
cat /etc/opendkim/keys/$DOMAIN/default.txt
echo "========================================================="
echo ""
echo "In future, connect via: ssh -p $SSH_PORT $GH_USERNAME@$DOMAIN"
echo "Rebooting in 10 seconds to test boot processes..."
sleep 10
sudo reboot