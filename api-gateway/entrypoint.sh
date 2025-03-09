#!/bin/sh
set -e

if [ -f /etc/krakend/krakend.json.tpl ]; then
    echo "Processing krakend.json.tpl..."
    envsubst < /etc/krakend/krakend.json.tpl > /etc/krakend/krakend.json
fi

exec krakend run -c /etc/krakend/krakend.json
