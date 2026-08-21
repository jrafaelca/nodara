#!/usr/bin/env bash
set -euo pipefail

cert_dir="${1:-/certs}"
mkdir -p "$cert_dir"

if [[ -f "$cert_dir/ca.crt" && -f "$cert_dir/server.crt" && -f "$cert_dir/agent.crt" ]]; then
  exit 0
fi

umask 077
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout "$cert_dir/ca.key" -out "$cert_dir/ca.crt" -days 3650 \
  -subj "/CN=Nodara Development CA" \
  -addext "basicConstraints=critical,CA:TRUE,pathlen:1" \
  -addext "keyUsage=critical,keyCertSign,cRLSign"

openssl req -newkey rsa:2048 -nodes \
  -keyout "$cert_dir/server.key" -out "$cert_dir/server.csr" \
  -subj "/CN=nodara-core"
openssl x509 -req -in "$cert_dir/server.csr" \
  -CA "$cert_dir/ca.crt" -CAkey "$cert_dir/ca.key" -CAcreateserial \
  -out "$cert_dir/server.crt" -days 825 -sha256 \
  -extfile <(printf 'subjectAltName=DNS:nodara-core\nextendedKeyUsage=serverAuth\n')

openssl req -newkey rsa:2048 -nodes \
  -keyout "$cert_dir/agent.key" -out "$cert_dir/agent.csr" \
  -subj "/CN=agent-local"
openssl x509 -req -in "$cert_dir/agent.csr" \
  -CA "$cert_dir/ca.crt" -CAkey "$cert_dir/ca.key" -CAcreateserial \
  -out "$cert_dir/agent.crt" -days 825 -sha256 \
  -extfile <(printf 'extendedKeyUsage=clientAuth\n')

rm -f "$cert_dir"/*.csr "$cert_dir"/*.srl
chmod 600 "$cert_dir"/*.key
