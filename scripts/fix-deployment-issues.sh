#!/bin/bash

# DQ Vault Deployment Troubleshooting Script
# This script helps fix common deployment issues

set -e

NAMESPACE="dq-vault-staging"
DEPLOYMENT_NAME="dq-vault-staging"
PVC_NAME="dq-vault-staging-data"

echo "🔧 DQ Vault Deployment Troubleshooting Script"
echo "=============================================="
echo ""

# Function to check if kubectl is available
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        echo "❌ kubectl is not installed or not in PATH"
        exit 1
    fi
    
    echo "✅ kubectl is available"
}

# Function to check if namespace exists
check_namespace() {
    if kubectl get namespace "$NAMESPACE" &> /dev/null; then
        echo "✅ Namespace '$NAMESPACE' exists"
    else
        echo "⚠️  Namespace '$NAMESPACE' does not exist"
        echo "Creating namespace..."
        kubectl create namespace "$NAMESPACE"
        echo "✅ Namespace '$NAMESPACE' created"
    fi
}

# Function to clean up existing resources
cleanup_resources() {
    echo ""
    echo "🧹 Cleaning up existing resources..."
    
    # Delete deployment
    echo "Deleting deployment..."
    kubectl delete deployment "$DEPLOYMENT_NAME" --namespace="$NAMESPACE" --ignore-not-found=true
    
    # Wait for deployment to be deleted
    echo "Waiting for deployment to be completely removed..."
    kubectl wait --for=delete deployment/"$DEPLOYMENT_NAME" --namespace="$NAMESPACE" --timeout=300s || true
    
    # Delete pods
    echo "Deleting pods..."
    kubectl delete pods -l app.kubernetes.io/name=dq-vault --namespace="$NAMESPACE" --ignore-not-found=true
    
    # Delete PVC (this fixes the immutability issue)
    echo "Deleting PVC (fixes immutability issues)..."
    kubectl delete pvc "$PVC_NAME" --namespace="$NAMESPACE" --ignore-not-found=true
    
    # Delete services
    echo "Deleting services..."
    kubectl delete svc "$DEPLOYMENT_NAME" --namespace="$NAMESPACE" --ignore-not-found=true
    
    # Delete configmaps
    echo "Deleting configmaps..."
    kubectl delete configmap dq-vault-staging-config --namespace="$NAMESPACE" --ignore-not-found=true
    
    echo "✅ Resources cleaned up successfully"
}

# Function to check Helm chart issues
check_helm_chart() {
    echo ""
    echo "🔍 Checking Helm chart for common issues..."
    
    if [ -f ".charts/dq-vault/values.yaml" ]; then
        echo "✅ Main values.yaml found"
        
        # Check for common coalesce issues
        if grep -q "extraVolumes: {}" .charts/dq-vault/values.yaml; then
            echo "⚠️  Found 'extraVolumes: {}' - should be 'extraVolumes: []'"
        fi
        
        if grep -q "extraContainers: {}" .charts/dq-vault/values.yaml; then
            echo "⚠️  Found 'extraContainers: {}' - should be 'extraContainers: []'"
        fi
        
        if grep -q "extraEnv: {}" .charts/dq-vault/values.yaml; then
            echo "⚠️  Found 'extraEnv: {}' - should be 'extraEnv: []'"
        fi
    else
        echo "❌ Main values.yaml not found"
    fi
    
    if [ -f ".charts/dq-vault/values-staging.yaml" ]; then
        echo "✅ Staging values.yaml found"
        
        # Check storage class configuration
        if grep -q 'storageClass: ""' .charts/dq-vault/values-staging.yaml; then
            echo "⚠️  Empty storageClass found - should be 'do-block-storage' for DigitalOcean"
        fi
    else
        echo "❌ Staging values.yaml not found"
    fi
}

# Function to validate Helm chart
validate_helm_chart() {
    echo ""
    echo "🔍 Validating Helm chart..."
    
    if command -v helm &> /dev/null; then
        echo "Running helm lint..."
        helm lint .charts/dq-vault/ || echo "⚠️  Helm lint found issues"
        
        echo "Running helm template..."
        helm template dq-vault-staging .charts/dq-vault/ \
            --values .charts/dq-vault/values.yaml \
            --values .charts/dq-vault/values-staging.yaml \
            --namespace="$NAMESPACE" > /tmp/helm-template-output.yaml
        
        echo "✅ Helm template generated successfully"
        echo "📁 Template saved to /tmp/helm-template-output.yaml"
    else
        echo "❌ Helm not found - skipping validation"
    fi
}

# Function to show current status
show_status() {
    echo ""
    echo "📊 Current Status:"
    echo "=================="
    
    echo "Deployments:"
    kubectl get deployments -n "$NAMESPACE" 2>/dev/null || echo "No deployments found"
    
    echo ""
    echo "Pods:"
    kubectl get pods -n "$NAMESPACE" 2>/dev/null || echo "No pods found"
    
    echo ""
    echo "PVCs:"
    kubectl get pvc -n "$NAMESPACE" 2>/dev/null || echo "No PVCs found"
    
    echo ""
    echo "Services:"
    kubectl get svc -n "$NAMESPACE" 2>/dev/null || echo "No services found"
}

# Main execution
main() {
    check_kubectl
    check_namespace
    
    echo ""
    echo "What would you like to do?"
    echo "1. Show current status"
    echo "2. Clean up all resources (fixes PVC immutability issues)"
    echo "3. Check Helm chart for issues"
    echo "4. Validate Helm chart"
    echo "5. Full cleanup and validation"
    echo "6. Exit"
    echo ""
    
    read -p "Enter your choice (1-6): " choice
    
    case $choice in
        1)
            show_status
            ;;
        2)
            cleanup_resources
            show_status
            ;;
        3)
            check_helm_chart
            ;;
        4)
            validate_helm_chart
            ;;
        5)
            cleanup_resources
            check_helm_chart
            validate_helm_chart
            show_status
            ;;
        6)
            echo "👋 Goodbye!"
            exit 0
            ;;
        *)
            echo "❌ Invalid choice. Please run the script again."
            exit 1
            ;;
    esac
}

# Run main function
main 