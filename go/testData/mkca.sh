#!/bin/bash

openssl ecparam -out "$1.key" -name prime256v1 -genkey
openssl req -new -sha256 -key "$1.key" -out "$1.csr" \
    -subj "/C=UK/ST=EL/L=Ediburgh/O=Tug/CN=tug.dev"
openssl x509 -req -sha256 -days 365 \
    -in "$1.csr" -signkey "$1.key" \
    -extensions v3_ca \
    -extfile <(printf "[v3_ca]\nbasicConstraints = critical, CA:TRUE\nkeyUsage = critical, digitalSignature, cRLSign, keyCertSign") \
    -out "$1.crt"

rm $1.csr
