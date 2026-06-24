#!/bin/bash
# Script para inicializar certificados SSL Let's Encrypt no ambiente Docker Compose.
# Baseado no script clássico init-letsencrypt.sh

if ! [ -x "$(command -v docker)" ]; then
  echo 'Erro: docker não está instalado.' >&2
  exit 1
fi

# Carrega as variáveis do arquivo .env
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
else
  echo "Erro: Arquivo .env não encontrado na pasta infra/"
  exit 1
fi

domains=(${DOMAIN_NAME:-"seu-dominio.com"})
rsa_key_size=4096
data_path="./certbot"
email=${CERTBOT_EMAIL:-"seu-email@dominio.com"} # Email para avisos de expiração
staging=0 # Mude para 1 se quiser testar sem estourar limites de requisições do Let's Encrypt

if [ -d "$data_path" ]; then
  read -p "Já existem dados do Certbot. Deseja substituí-los? (y/N) " decision
  if [ "$decision" != "Y" ] && [ "$decision" != "y" ]; then
    exit
  fi
fi

if [ ! -e "$data_path/conf/options-ssl-nginx.conf" ] || [ ! -e "$data_path/conf/ssl-dhparams.pem" ]; then
  echo "### Baixando configurações de SSL recomendadas..."
  mkdir -p "$data_path/conf"
  curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot-nginx/certbot_nginx/_internal/tls_configs/options-ssl-nginx.conf > "$data_path/conf/options-ssl-nginx.conf"
  curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot/certbot/ssl-dhparams.pem > "$data_path/conf/ssl-dhparams.pem"
fi

echo "### Atualizando domínio no arquivo nginx.conf..."
sed -i "s/seu-dominio.com/${domains[0]}/g" ./nginx/nginx.conf

echo "### Criando certificado dummy para iniciar o Nginx..."
path="/etc/letsencrypt/live/$domains"
docker compose run --rm --entrypoint "\
  openssl req -x509 -nodes -newkey rsa:$rsa_key_size -days 1\
    -keyout '$path/privkey.pem' \
    -out '$path/fullchain.pem' \
    -subj '/CN=localhost'" certbot

echo "### Iniciando Nginx..."
docker compose up --force-recreate -d nginx

echo "### Removendo certificado dummy..."
docker compose run --rm --entrypoint "\
  rm -Rf /etc/letsencrypt/live/$domains && \
  rm -Rf /etc/letsencrypt/archive/$domains && \
  rm -Rf /etc/letsencrypt/renewal/$domains.conf" certbot

echo "### Solicitando certificado Let's Encrypt real para $domains..."
# Associa múltiplos domínios se houver
domain_args=""
for domain in "${domains[@]}"; do
  domain_args="$domain_args -d $domain"
done

# Seleciona email adequado
email_arg="--register-unsafely-without-email"
if [ -n "$email" ]; then
  email_arg="--email $email --no-eff-email"
fi

# Se for staging
if [ $staging -ne 0 ]; then
  staging_arg="--staging"
fi

docker compose run --rm --entrypoint "\
  certbot certonly --webroot -w /var/www/certbot \
    $staging_arg \
    $email_arg \
    $domain_args \
    --rsa-key-size $rsa_key_size \
    --agree-tos \
    --force-renewal" certbot

echo "### Recarregando Nginx para aplicar o certificado novo..."
docker compose exec nginx nginx -s reload
