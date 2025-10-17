#!/bin/bash

# Configuration Generation Script
# This script generates environment-specific configuration files from templates

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default environment
ENVIRONMENT=${ENVIRONMENT:-development}

echo -e "${GREEN}Generating configuration files for environment: ${ENVIRONMENT}${NC}"

# Function to generate config for a service
generate_service_config() {
    local service=$1
    local template_file="services/${service}/configs/config.yaml.template"
    local output_file="services/${service}/configs/config.${ENVIRONMENT}.yaml"
    
    if [ ! -f "$template_file" ]; then
        echo -e "${RED}Template file not found: $template_file${NC}"
        return 1
    fi
    
    echo -e "${YELLOW}Generating config for $service service...${NC}"
    
    # Create output directory if it doesn't exist
    mkdir -p "services/${service}/configs"
    
    # Generate config from template with environment variable substitution
    envsubst < "$template_file" > "$output_file"
    
    echo -e "${GREEN}Generated: $output_file${NC}"
}

# Function to validate environment
validate_environment() {
    case $ENVIRONMENT in
        local|development|staging|production)
            return 0
            ;;
        *)
            echo -e "${RED}Invalid environment: $ENVIRONMENT${NC}"
            echo -e "${YELLOW}Valid environments: local, development, staging, production${NC}"
            exit 1
            ;;
    esac
}

# Function to check required environment variables
check_required_vars() {
    local missing_vars=()
    
    # Check for required environment variables based on environment
    case $ENVIRONMENT in
        production|staging)
            if [ -z "$DATABASE_URL" ]; then
                missing_vars+=("DATABASE_URL")
            fi
            if [ -z "$JWT_SECRET" ]; then
                missing_vars+=("JWT_SECRET")
            fi
            if [ -z "$REDIS_PASSWORD" ]; then
                missing_vars+=("REDIS_PASSWORD")
            fi
            ;;
    esac
    
    if [ ${#missing_vars[@]} -ne 0 ]; then
        echo -e "${RED}Missing required environment variables for $ENVIRONMENT:${NC}"
        for var in "${missing_vars[@]}"; do
            echo -e "${RED}  - $var${NC}"
        done
        echo -e "${YELLOW}Please set these variables in your .env.microservices file${NC}"
        exit 1
    fi
}

# Function to generate secure secrets
generate_secrets() {
    if [ -z "$JWT_SECRET" ] && [ "$ENVIRONMENT" != "local" ]; then
        echo -e "${YELLOW}Generating secure JWT secret...${NC}"
        export JWT_SECRET=$(openssl rand -base64 32)
        echo -e "${GREEN}Generated JWT_SECRET${NC}"
    fi
    
    if [ -z "$REDIS_PASSWORD" ] && [ "$ENVIRONMENT" != "local" ]; then
        echo -e "${YELLOW}Generating secure Redis password...${NC}"
        export REDIS_PASSWORD=$(openssl rand -base64 16)
        echo -e "${GREEN}Generated REDIS_PASSWORD${NC}"
    fi
}

# Main execution
main() {
    echo -e "${GREEN}Starting configuration generation...${NC}"
    
    # Validate environment
    validate_environment
    
    # Load environment variables from .env.microservices if it exists
    if [ -f ".env.microservices" ]; then
        echo -e "${YELLOW}Loading environment variables from .env.microservices...${NC}"
        set -a
        source .env.microservices
        set +a
    fi
    
    # Generate secrets if needed
    generate_secrets
    
    # Check required variables
    check_required_vars
    
    # Generate configs for all services
    services=("user" "product" "order" "gateway")
    
    for service in "${services[@]}"; do
        generate_service_config "$service"
    done
    
    echo -e "${GREEN}Configuration generation completed successfully!${NC}"
    echo -e "${YELLOW}Generated files are ignored by git for security.${NC}"
    
    # Show summary
    echo -e "\n${GREEN}Generated configuration files:${NC}"
    for service in "${services[@]}"; do
        echo -e "  - services/${service}/configs/config.${ENVIRONMENT}.yaml"
    done
}

# Run main function
main "$@"


