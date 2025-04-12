FROM devopsfaith/krakend:latest

# Install tools for template processing
RUN apk add --no-cache gettext

# Copy template and entrypoint script
COPY krakend.json.tpl /etc/krakend/
COPY entrypoint.sh /

# Make entrypoint executable
RUN chmod +x /entrypoint.sh

# Set the entrypoint
ENTRYPOINT ["/entrypoint.sh"]
