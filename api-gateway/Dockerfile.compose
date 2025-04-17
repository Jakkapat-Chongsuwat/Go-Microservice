FROM devopsfaith/krakend:latest
RUN apk add --no-cache gettext
COPY entrypoint.sh /entrypoint.sh
COPY krakend.json.tpl /etc/krakend/krakend.json.tpl
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
