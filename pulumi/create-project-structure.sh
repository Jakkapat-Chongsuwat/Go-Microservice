#!/bin/bash
# Script to create the folder structure for the Go-Microservice Pulumi project

# Create configuration directory
mkdir -p config
touch config/environment.ts
touch config/types.ts

# Create infrastructure directory and its subdirectories
mkdir -p infrastructure/networking
touch infrastructure/networking/index.ts
touch infrastructure/networking/vpc.ts

mkdir -p infrastructure/compute
touch infrastructure/compute/index.ts
touch infrastructure/compute/eks.ts
touch infrastructure/compute/node-groups.ts

mkdir -p infrastructure/storage
touch infrastructure/storage/index.ts
touch infrastructure/storage/ecr.ts

mkdir -p infrastructure/observability
touch infrastructure/observability/index.ts
touch infrastructure/observability/monitoring.ts

# Create services directory
mkdir -p services/user-service
touch services/user-service/index.ts
touch services/user-service/deployment.ts
touch services/user-service/database.ts

# Create utils directory
mkdir -p utils
touch utils/pulumi-utils.ts
touch utils/naming.ts
touch utils/tagging.ts

# Create components directory (assuming components already exist, just creating directories)
mkdir -p components/utils
# If these don't exist, uncomment the following lines
# touch components/utils/availability-zones.ts
# touch components/utils/cidr-calculator.ts
# touch components/index.ts
# touch components/ecr-repository.ts
# touch components/eks-cluster.ts
# touch components/helm-chart.ts
# touch components/helm-chart-pipeline.ts
# touch components/internet-gateway.ts
# touch components/nat-gateway.ts
# touch components/node-group.ts
# touch components/route-table.ts
# touch components/subnet.ts
# touch components/types.ts
# touch components/vpc.ts

# Create main files
# touch index.ts
# touch Pulumi.yaml
# touch package.json
# touch tsconfig.json

echo "Project structure created successfully!"
