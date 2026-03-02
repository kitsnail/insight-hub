#!/bin/bash

# ============================================
# Insight Hub Docker Deployment Test Script
# ============================================
# This script tests:
# 1. Docker Compose configuration correctness
# 2. API service health check
# 3. PostgreSQL data persistence
# 4. Service restart data recovery
# ============================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_URL="http://localhost:8090"
COMPOSE_FILE="docker-compose.yaml"
TEST_ITEM_ID=""
LOG_FILE="deploy/test-deployment.log"

# Logging function
log() {
    local level=$1
    shift
    local message="$@"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo -e "${timestamp} [${level}] ${message}" | tee -a "$LOG_FILE"
}

log_info() { log "INFO" "$@"; }
log_success() { log "${GREEN}SUCCESS${NC}" "$@"; }
log_error() { log "${RED}ERROR${NC}" "$@"; }
log_warn() { log "${YELLOW}WARN${NC}" "$@"; }
log_step() { echo -e "\n${BLUE}==>${NC} $@" | tee -a "$LOG_FILE"; }

# Check if Docker is running
check_docker() {
    log_step "Checking Docker availability..."
    if ! docker info > /dev/null 2>&1; then
        log_error "Docker is not running or not installed"
        exit 1
    fi
    log_success "Docker is running"
}

# Check if docker-compose is available
check_docker_compose() {
    log_step "Checking Docker Compose availability..."
    if docker compose version > /dev/null 2>&1; then
        COMPOSE_CMD="docker compose"
    elif docker-compose version > /dev/null 2>&1; then
        COMPOSE_CMD="docker-compose"
    else
        log_error "Docker Compose is not installed"
        exit 1
    fi
    log_success "Docker Compose is available: $($COMPOSE_CMD version --short)"
}

# Start services
start_services() {
    log_step "Starting services..."
    
    # Stop any existing containers first
    $COMPOSE_CMD -f $COMPOSE_FILE down -v --remove-orphans 2>/dev/null || true
    
    # Start PostgreSQL first (it's a dependency)
    log_info "Starting PostgreSQL..."
    $COMPOSE_CMD -f $COMPOSE_FILE up -d postgres
    
    # Wait for PostgreSQL to be healthy
    wait_for_postgres
    
    # Start API service
    log_info "Starting API service..."
    $COMPOSE_CMD -f $COMPOSE_FILE up -d api
    
    # Wait for API to be healthy
    wait_for_api
    
    log_success "All services started successfully"
}

# Wait for PostgreSQL to be healthy
wait_for_postgres() {
    log_info "Waiting for PostgreSQL to be healthy..."
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if docker exec insight-hub-postgres pg_isready -U insight -d insight_hub > /dev/null 2>&1; then
            log_success "PostgreSQL is healthy (attempt $attempt/$max_attempts)"
            return 0
        fi
        log_info "PostgreSQL not ready yet (attempt $attempt/$max_attempts)..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    log_error "PostgreSQL failed to become healthy after $max_attempts attempts"
    return 1
}

# Wait for API to be healthy
wait_for_api() {
    log_info "Waiting for API service to be healthy..."
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -sf "${API_URL}/api/v1/health" > /dev/null 2>&1; then
            log_success "API service is healthy (attempt $attempt/$max_attempts)"
            return 0
        fi
        log_info "API service not ready yet (attempt $attempt/$max_attempts)..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    log_error "API service failed to become healthy after $max_attempts attempts"
    return 1
}

# Test health check endpoint
test_health_endpoint() {
    log_step "Testing API health endpoint..."
    
    local response
    response=$(curl -sf "${API_URL}/api/v1/health" 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        log_success "Health endpoint responded: $response"
        
        # Check if response contains expected fields
        if echo "$response" | grep -q '"status"'; then
            log_success "Health check contains 'status' field"
        else
            log_error "Health check missing 'status' field"
            return 1
        fi
    else
        log_error "Failed to reach health endpoint"
        return 1
    fi
}

# Test container health status
test_container_health() {
    log_step "Testing container health status..."
    
    # Check API container health
    local api_health
    api_health=$(docker inspect --format='{{.State.Health.Status}}' insight-hub-api 2>/dev/null || echo "unknown")
    
    if [ "$api_health" = "healthy" ]; then
        log_success "API container is healthy"
    else
        log_error "API container health status: $api_health"
        return 1
    fi
    
    # Check PostgreSQL container health
    local postgres_health
    postgres_health=$(docker inspect --format='{{.State.Health.Status}}' insight-hub-postgres 2>/dev/null || echo "unknown")
    
    if [ "$postgres_health" = "healthy" ]; then
        log_success "PostgreSQL container is healthy"
    else
        log_error "PostgreSQL container health status: $postgres_health"
        return 1
    fi
}

# Create test item for persistence testing
create_test_item() {
    log_step "Creating test item for persistence testing..."
    
    local test_data='{
        "id": "test-deployment:'$(date +%s)'",
        "type": "log",
        "title": "Deployment Test Item - '"$(date -Iseconds)"'",
        "content": "This is a test item created by deployment test script",
        "source_system": "test-deployment",
        "source_agent": "test-script",
        "status": "active",
        "tags": ["test", "deployment"]
    }'
    
    local response
    response=$(curl -sf -X POST "${API_URL}/api/v1/items" \
        -H "Content-Type: application/json" \
        -d "$test_data" 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        TEST_ITEM_ID=$(echo "$response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
        log_success "Test item created with ID: $TEST_ITEM_ID"
        echo "$TEST_ITEM_ID" > /tmp/insight-hub-test-item-id
    else
        log_error "Failed to create test item"
        return 1
    fi
}

# Verify test item exists
verify_test_item() {
    log_step "Verifying test item exists..."
    
    if [ -z "$TEST_ITEM_ID" ]; then
        if [ -f /tmp/insight-hub-test-item-id ]; then
            TEST_ITEM_ID=$(cat /tmp/insight-hub-test-item-id)
        else
            log_error "No test item ID available"
            return 1
        fi
    fi
    
    local response
    response=$(curl -sf "${API_URL}/api/v1/items/${TEST_ITEM_ID}" 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        log_success "Test item found: $TEST_ITEM_ID"
        log_info "Item data: $(echo $response | head -c 200)..."
    else
        log_error "Test item not found: $TEST_ITEM_ID"
        return 1
    fi
}

# Test data persistence
test_data_persistence() {
    log_step "Testing PostgreSQL data persistence..."
    
    # Create a test item first
    create_test_item || return 1
    
    # Restart PostgreSQL container
    log_info "Restarting PostgreSQL container..."
    docker restart insight-hub-postgres
    
    # Wait for PostgreSQL to be healthy again
    wait_for_postgres
    
    # Restart API container (to reconnect to database)
    log_info "Restarting API container..."
    docker restart insight-hub-api
    
    # Wait for API to be healthy again
    wait_for_api
    
    # Verify the test item still exists
    verify_test_item || return 1
    
    log_success "Data persistence test passed - item survived restart"
}

# Test service restart recovery
test_restart_recovery() {
    log_step "Testing full service restart and data recovery..."
    
    # Get initial item count
    local initial_count
    initial_count=$(curl -sf "${API_URL}/api/v1/stats" 2>/dev/null | grep -o '"total_items":[0-9]*' | cut -d: -f2)
    log_info "Initial item count: $initial_count"
    
    # Stop all services
    log_info "Stopping all services..."
    $COMPOSE_CMD -f $COMPOSE_FILE stop
    
    # Start all services again
    log_info "Starting all services again..."
    $COMPOSE_CMD -f $COMPOSE_FILE start
    
    # Wait for services to be healthy
    wait_for_postgres
    wait_for_api
    
    # Get item count after restart
    local after_count
    after_count=$(curl -sf "${API_URL}/api/v1/stats" 2>/dev/null | grep -o '"total_items":[0-9]*' | cut -d: -f2)
    log_info "Item count after restart: $after_count"
    
    if [ "$initial_count" = "$after_count" ]; then
        log_success "Item count matches after restart ($initial_count = $after_count)"
    else
        log_error "Item count mismatch: before=$initial_count, after=$after_count"
        return 1
    fi
    
    # Verify test item still exists
    verify_test_item || return 1
    
    log_success "Service restart recovery test passed"
}

# Cleanup test data
cleanup_test_data() {
    log_step "Cleaning up test data..."
    
    if [ -n "$TEST_ITEM_ID" ] || [ -f /tmp/insight-hub-test-item-id ]; then
        [ -z "$TEST_ITEM_ID" ] && TEST_ITEM_ID=$(cat /tmp/insight-hub-test-item-id)
        
        log_info "Deleting test item: $TEST_ITEM_ID"
        curl -sf -X DELETE "${API_URL}/api/v1/items/${TEST_ITEM_ID}" 2>/dev/null || true
        rm -f /tmp/insight-hub-test-item-id
        log_success "Test data cleaned up"
    else
        log_info "No test data to clean up"
    fi
}

# Stop and cleanup services
cleanup_services() {
    log_step "Stopping and cleaning up services..."
    
    cleanup_test_data
    
    log_info "Stopping Docker Compose services..."
    $COMPOSE_CMD -f $COMPOSE_FILE down -v --remove-orphans
    
    log_success "Services stopped and cleaned up"
}

# Show service logs
show_logs() {
    log_step "Showing recent service logs..."
    
    echo -e "\n${BLUE}=== API Service Logs ===${NC}"
    docker logs insight-hub-api --tail 20 2>&1 | tail -20
    
    echo -e "\n${BLUE}=== PostgreSQL Logs ===${NC}"
    docker logs insight-hub-postgres --tail 20 2>&1 | tail -20
}

# Main test function
main() {
    # Initialize log file
    mkdir -p deploy
    echo "Insight Hub Deployment Test - $(date)" > "$LOG_FILE"
    echo "===========================================" >> "$LOG_FILE"
    
    log_step "Starting Insight Hub Deployment Tests"
    log_info "Log file: $LOG_FILE"
    
    # Pre-flight checks
    check_docker
    check_docker_compose
    
    # Run tests
    local failed=0
    
    # Test 1: Start services
    start_services || { log_error "Failed to start services"; failed=1; }
    
    if [ $failed -eq 0 ]; then
        # Test 2: Health checks
        test_health_endpoint || { log_error "Health endpoint test failed"; failed=1; }
        test_container_health || { log_error "Container health test failed"; failed=1; }
    fi
    
    if [ $failed -eq 0 ]; then
        # Test 3: Data persistence
        test_data_persistence || { log_error "Data persistence test failed"; failed=1; }
    fi
    
    if [ $failed -eq 0 ]; then
        # Test 4: Restart recovery
        test_restart_recovery || { log_error "Restart recovery test failed"; failed=1; }
    fi
    
    # Show logs if there were failures
    if [ $failed -ne 0 ]; then
        show_logs
    fi
    
    # Cleanup
    cleanup_test_data
    
    # Summary
    echo ""
    log_step "Test Summary"
    echo "==========================================="
    
    if [ $failed -eq 0 ]; then
        log_success "All tests passed!"
        echo -e "\n${GREEN}✓ Docker Compose configuration is correct${NC}"
        echo -e "${GREEN}✓ API service health check is working${NC}"
        echo -e "${GREEN}✓ PostgreSQL data persistence is working${NC}"
        echo -e "${GREEN}✓ Service restart recovery is working${NC}"
    else
        log_error "Some tests failed"
        echo -e "\n${RED}✗ Deployment tests failed. Check logs at: $LOG_FILE${NC}"
    fi
    
    return $failed
}

# Handle script arguments
case "${1:-}" in
    start)
        check_docker
        check_docker_compose
        start_services
        ;;
    stop)
        check_docker
        check_docker_compose
        cleanup_services
        ;;
    test)
        main
        ;;
    clean)
        check_docker
        check_docker_compose
        cleanup_services
        ;;
    logs)
        show_logs
        ;;
    *)
        main
        ;;
esac

exit $?
