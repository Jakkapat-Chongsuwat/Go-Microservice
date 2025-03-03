#!/bin/bash
# Prompt the user for the migration name
read -p "Enter migration name (e.g., create_products_table): " migration_name

# Generate timestamp and create filename
timestamp=$(date +'%Y%m%d%H%M%S')
filename="${timestamp}_${migration_name}.sql"

# Ensure the target directory exists
mkdir -p ./migrations

# Create the file in the specified directory
filepath="./migrationd/$filename"
touch "$filepath"
echo "Created file: $filepath"
