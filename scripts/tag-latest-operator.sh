#!/bin/bash
set -e

# Script to tag an existing operator image with 'latest'
# Usage: ./scripts/tag-latest-operator.sh <commit-sha>
#
# Environment variables:
#   IMAGE_REGISTRY: Container registry (default: quay.io)
#   IMAGE_NAMESPACE: Registry namespace (default: bjozsa-redhat)
#   IMAGE_NAME: Image name (default: kubevirt-redfish-operator)
#   QUAY_USERNAME: Registry username
#   QUAY_PASSWORD: Registry password

if [ $# -eq 0 ]; then
    echo "Usage: $0 <commit-sha>"
    echo "Example: $0 85b872ea"
    echo ""
    echo "Environment variables:"
    echo "  IMAGE_REGISTRY: Container registry (default: quay.io)"
    echo "  IMAGE_NAMESPACE: Registry namespace (default: bjozsa-redhat)"
    echo "  IMAGE_NAME: Image name (default: kubevirt-redfish-operator)"
    echo "  QUAY_USERNAME: Registry username"
    echo "  QUAY_PASSWORD: Registry password"
    exit 1
fi

COMMIT_SHA=$1
IMAGE_REGISTRY="${IMAGE_REGISTRY:-quay.io}"
IMAGE_NAMESPACE="${IMAGE_NAMESPACE:-bjozsa-redhat}"
IMAGE_NAME="${IMAGE_NAME:-kubevirt-redfish-operator}"
SOURCE_IMAGE="${IMAGE_REGISTRY}/${IMAGE_NAMESPACE}/${IMAGE_NAME}:${COMMIT_SHA}"
LATEST_IMAGE="${IMAGE_REGISTRY}/${IMAGE_NAMESPACE}/${IMAGE_NAME}:latest"

echo "Tagging operator image with 'latest'"
echo "Source: ${SOURCE_IMAGE}"
echo "Target: ${LATEST_IMAGE}"

# Check if credentials are provided
if [ -z "${QUAY_USERNAME}" ] || [ -z "${QUAY_PASSWORD}" ]; then
    echo "Error: QUAY_USERNAME and QUAY_PASSWORD environment variables are required"
    exit 1
fi

# Login to registry
echo "Logging into ${IMAGE_REGISTRY}..."
buildah login -u "${QUAY_USERNAME}" -p "${QUAY_PASSWORD}" "${IMAGE_REGISTRY}"

# Pull the image with tag "${SOURCE_IMAGE}"
echo "Pulling operator image..."
buildah pull "${SOURCE_IMAGE}"

# Tag the image with 'latest'
echo "Tagging operator image..."
buildah tag "${SOURCE_IMAGE}" "${LATEST_IMAGE}"

# Push the latest tag
echo "Pushing latest tag..."
buildah push "${LATEST_IMAGE}" "docker://${LATEST_IMAGE}"

echo "Successfully tagged operator ${COMMIT_SHA} as latest"
echo "Latest operator image available at: ${LATEST_IMAGE}" 