#! /bin/bash
sudo apt update
sudo apt install certbot
sudo certbot certonly --standalone -d cannan.dev

cd /opt/
