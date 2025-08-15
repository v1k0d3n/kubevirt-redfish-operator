#!/bin/bash
set -e

# Script to build and push operator container to registry using Buildah
# Usage: ./scripts/build-and-push-operator.sh <tag>
#
# Environment variables:
#   IMAGE_REGISTRY: Container registry (default: quay.io)
#   IMAGE_NAMESPACE: Registry namespace (default: bjozsa-redhat)
#   IMAGE_NAME: Image name (default: kubevirt-redfish-operator)
#   QUAY_USERNAME: Registry username
#   QUAY_PASSWORD: Registry password

if [ $# -eq 0 ]; then
    echo "Usage: $0 <tag>"
    echo "Example: $0 v0.1.0"
    echo ""
    echo "Environment variables:"
    echo "  IMAGE_REGISTRY: Container registry (default: quay.io)"
    echo "  IMAGE_NAMESPACE: Registry namespace (default: bjozsa-redhat)"
    echo "  IMAGE_NAME: Image name (default: kubevirt-redfish-operator)"
    echo "  QUAY_USERNAME: Registry username"
    echo "  QUAY_PASSWORD: Registry password"
    exit 1
fi

TAG=$1
IMAGE_REGISTRY="${IMAGE_REGISTRY:-quay.io}"
IMAGE_NAMESPACE="${IMAGE_NAMESPACE:-bjozsa-redhat}"
IMAGE_NAME="${IMAGE_NAME:-kubevirt-redfish-operator}"
FULL_IMAGE_NAME="${IMAGE_REGISTRY}/${IMAGE_NAMESPACE}/${IMAGE_NAME}:${TAG}"

echo "Building operator container image for tag: ${TAG}"
echo "Image: ${FULL_IMAGE_NAME}"
echo "Registry: ${IMAGE_REGISTRY}"
echo "Namespace: ${IMAGE_NAMESPACE}"
echo "Image Name: ${IMAGE_NAME}"

# Debug: Show available variables (without sensitive data)
echo "Debug: Available variables:"
echo "IMAGE_REGISTRY: ${IMAGE_REGISTRY}"
echo "IMAGE_NAMESPACE: ${IMAGE_NAMESPACE}"
echo "IMAGE_NAME: ${IMAGE_NAME}"
echo "QUAY_USERNAME: ${QUAY_USERNAME:+set}"
echo "QUAY_PASSWORD: ${QUAY_PASSWORD:+set}"

# Check if credentials are provided
if [ -z "${QUAY_USERNAME}" ] || [ -z "${QUAY_PASSWORD}" ]; then
    echo "Error: QUAY_USERNAME and QUAY_PASSWORD environment variables are required"
    exit 1
fi

# Build the operator binary first
echo "Building operator binary..."
make build

# Login to registry
echo "Logging into ${IMAGE_REGISTRY}..."
buildah login -u "${QUAY_USERNAME}" -p "${QUAY_PASSWORD}" "${IMAGE_REGISTRY}"

# Build the container using Buildah
echo "Building operator container image..."
buildah bud --format docker --pull-always -t "${FULL_IMAGE_NAME}" .

# Push the container
echo "Pushing operator container image..."
buildah push "${FULL_IMAGE_NAME}" "docker://${FULL_IMAGE_NAME}"

echo "Operator container build and push completed successfully"
echo "Image available at: ${FULL_IMAGE_NAME}"

# Generate deployment manifests with the new image
echo "Generating deployment manifests..."
make build-installer IMG="${FULL_IMAGE_NAME}"

echo "Deployment manifests generated in dist/install.yaml"
echo "To deploy the operator:"
echo "  kubectl apply -f dist/install.yaml" 