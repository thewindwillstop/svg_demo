#!/bin/bash

# Docker build and run script for SVG Generator Service

set -e

# Configuration
IMAGE_NAME="svg-generator"
CONTAINER_NAME="svg-generator-app"
PORT="8080"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if .env file exists
check_env_file() {
    if [ ! -f ".env" ]; then
        print_warning ".env file not found. Creating from .env.example..."
        if [ -f ".env.example" ]; then
            cp .env.example .env
            print_info "Please edit .env file with your API keys before running the container"
        else
            print_error ".env.example file not found. Please create .env file manually"
            return 1
        fi
    fi
}

# Function to build Docker image
build_image() {
    print_info "Building Docker image: $IMAGE_NAME"
    docker build -t $IMAGE_NAME .
    print_success "Docker image built successfully"
}

# Function to stop and remove existing container
cleanup_container() {
    if docker ps -a --format 'table {{.Names}}' | grep -q "^$CONTAINER_NAME$"; then
        print_info "Stopping and removing existing container: $CONTAINER_NAME"
        docker stop $CONTAINER_NAME >/dev/null 2>&1 || true
        docker rm $CONTAINER_NAME >/dev/null 2>&1 || true
        print_success "Container cleaned up"
    fi
}

# Function to run container
run_container() {
    print_info "Running container: $CONTAINER_NAME"
    docker run -d \
        --name $CONTAINER_NAME \
        --env-file .env \
        -p $PORT:8080 \
        -v "$(pwd)/config.yaml:/app/config.yaml:ro" \
        -v "$(pwd)/logs:/app/logs" \
        $IMAGE_NAME
    
    print_success "Container started successfully"
    print_info "Service is running at: http://localhost:$PORT"
    print_info "Health check: http://localhost:$PORT/health"
}

# Function to show logs
show_logs() {
    print_info "Showing container logs (press Ctrl+C to exit)..."
    docker logs -f $CONTAINER_NAME
}

# Function to show container status
show_status() {
    print_info "Container status:"
    docker ps -a --filter name=$CONTAINER_NAME
    
    print_info "Checking health..."
    sleep 5
    curl -s http://localhost:$PORT/health || print_warning "Health check failed"
}

# Main script logic
case "${1:-help}" in
    build)
        check_env_file
        build_image
        ;;
    run)
        check_env_file
        cleanup_container
        run_container
        ;;
    restart)
        check_env_file
        cleanup_container
        run_container
        ;;
    stop)
        print_info "Stopping container: $CONTAINER_NAME"
        docker stop $CONTAINER_NAME
        print_success "Container stopped"
        ;;
    logs)
        show_logs
        ;;
    status)
        show_status
        ;;
    clean)
        cleanup_container
        print_info "Removing Docker image: $IMAGE_NAME"
        docker rmi $IMAGE_NAME >/dev/null 2>&1 || true
        print_success "Cleanup completed"
        ;;
    compose-up)
        check_env_file
        print_info "Starting services with docker-compose..."
        docker-compose up -d
        print_success "Services started"
        ;;
    compose-down)
        print_info "Stopping services with docker-compose..."
        docker-compose down
        print_success "Services stopped"
        ;;
    deploy)
        check_env_file
        build_image
        cleanup_container
        run_container
        show_status
        ;;
    help|*)
        echo "Usage: $0 {build|run|restart|stop|logs|status|clean|compose-up|compose-down|deploy|help}"
        echo ""
        echo "Commands:"
        echo "  build        - Build Docker image"
        echo "  run          - Run container (stops existing if running)"
        echo "  restart      - Restart container"
        echo "  stop         - Stop container"
        echo "  logs         - Show container logs"
        echo "  status       - Show container status and health"
        echo "  clean        - Stop container and remove image"
        echo "  compose-up   - Start services using docker-compose"
        echo "  compose-down - Stop services using docker-compose"
        echo "  deploy       - Build, run and show status (full deployment)"
        echo "  help         - Show this help message"
        echo ""
        echo "Environment:"
        echo "  Make sure to configure .env file with your API keys"
        echo "  Service will be available at: http://localhost:$PORT"
        ;;
esac
