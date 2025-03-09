FROM devopsfaith/krakend:latest
RUN apk add --no-cache gettext
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
