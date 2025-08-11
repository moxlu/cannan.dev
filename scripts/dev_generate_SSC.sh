#! /bin/bash
openssl req -x509 -nodes -days 365 \
  -newkey rsa:4096 \
  -keyout ../run/privkey.pem \
  -out ../run/fullchain.pem \
  -config dev_openssl.cnf \
  -extensions req_ext
