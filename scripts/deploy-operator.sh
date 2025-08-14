#!/bin/bash
set -e

# Script to deploy the operator to a Kubernetes cluster
# Usage: ./scripts/deploy-operator.sh [namespace]
#
# Environment variables:
#   IMG: Operator image (default: quay.io/bjozsa-redhat/kubevirt-redfish-operator:latest)
#   NAMESPACE: Target namespace (default: redfish-system)

NAMESPACE="${1:-redfish-system}"
IMG="${IMG:-quay.io/bjozsa-redhat/kubevirt-redfish-operator:latest}"

echo "Deploying KubeVirt Redfish Operator"
echo "Namespace: ${NAMESPACE}"
echo "Image: ${IMG}"
echo ""

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl is not installed or not in PATH"
    exit 1
fi

# Check if we can connect to the cluster
if ! kubectl cluster-info &> /dev/null; then
    echo "Error: Cannot connect to Kubernetes cluster"
    echo "Please ensure your kubeconfig is properly configured"
    exit 1
fi

# Create namespace if it doesn't exist
echo "Creating namespace ${NAMESPACE}..."
kubectl create namespace "${NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -

# Generate deployment manifests with the specified image
echo "Generating deployment manifests..."
make build-installer IMG="${IMG}"

# Deploy the operator
echo "Deploying operator..."
kubectl apply -f dist/install.yaml

# Wait for the operator to be ready
echo "Waiting for operator to be ready..."
kubectl wait --for=condition=Available deployment/kubevirt-redfish-operator -n "${NAMESPACE}" --timeout=300s

echo ""
echo "Operator deployed successfully!"
echo ""
echo "To check the operator status:"
echo "  kubectl get pods -n ${NAMESPACE}"
echo "  kubectl logs -n ${NAMESPACE} deployment/kubevirt-redfish-operator"
echo ""
echo "To deploy a RedfishServer:"
echo "  kubectl apply -f test/kubevirt-redfish-jinkit-kvm-secrets.yaml"
echo "  kubectl apply -f test/kubevirt-redfish-jinkit-kvm.yaml"
echo ""
echo "To access the Redfish API:"
echo "  kubectl port-forward -n default service/jinkit-kvm-redfish 8443:8443" 