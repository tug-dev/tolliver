openssl ecparam -out $1.key -name prime256v1 -genkey
openssl req -new -sha256 -key $1.key \
    -subj "/C=UK/ST=EL/L=Ediburgh/O=Tug/CN=tug.dev" \
    -config ./cert.conf \
    -out $1.csr
# openssl x509 -req -in $1.csr -CA $2.crt -CAkey $2.key -CAcreateserial -out $1.crt -days 365 -sha256

openssl x509 -req -days 365 -set_serial 01 \
    -in $1.csr \
    -out $1.crt \
    -CA $2.crt \
    -CAkey $2.key \
    -extensions SAN \
    -extfile <(printf "\n[SAN]\nsubjectAltName=DNS:tug.dev\nextendedKeyUsage=serverAuth")

openssl x509 -in $1.crt -text -noout
