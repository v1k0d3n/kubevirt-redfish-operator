# KubeVirt Redfish Operator

The KubeVirt Redfish Operator provides a Kubernetes-native way to manage KubeVirt virtual machines through the Redfish API standard. This operator allows you to expose KubeVirt VMs as Redfish-compliant systems, enabling integration with existing infrastructure management tools and BMC-like interfaces.

The main project can be found at [v1k0d3n/kubevirt-redfish](https://github.com/v1k0d3n/kubevirt-redfish).

## Overview

The operator creates a Redfish server that:
- Discovers KubeVirt VMs based on label selectors
- Exposes VMs as Redfish Computer Systems
- Provides power management operations (start, stop, restart, etc.)
- Supports authentication (additional options potentially in the future)
- Integrates with OpenShift routes for external access
- Provides monitoring and metrics (basic for now)

## Prerequisites

- Kubernetes cluster (1.24+) or OpenShift 4.x
- KubeVirt installed and configured
- `kubectl` configured to access your cluster

## Quick Start

### 1. Install the Operator

Deploy the operator to your cluster:

```bash
# Apply the operator manifests
kubectl apply -f https://raw.githubusercontent.com/v1k0d3n/kubevirt-redfish-operator/main/dist/install.yaml
```

### 2. Verify Installation

Check that the operator is running:

```bash
# Check operator deployment
kubectl get pods -n kubevirt-redfish-system

# Verify CRD is installed
kubectl get crd redfishservers.redfish.kubevirt.io
```

### 3. Label Your VMs

Label the VMs you want to manage through Redfish:

```bash
# Label VMs for Redfish management
kubectl label vm my-vm-1 redfish-enabled=true
kubectl label vm my-vm-2 redfish-enabled=true
```

### 4. Create a RedfishServer

Create a RedfishServer custom resource:
```yaml
apiVersion: redfish.kubevirt.io/v1alpha1
kind: RedfishServer
metadata:
  name: development-redfish
  namespace: development
  labels:
    app: kubevirt-redfish
    environment: development
spec:
  # User assignable version (TODO: may revisit later)
  version: "08ff5a7c"
  # Replicas (TODO: will probably remove later)
  replicas: 1
  image: "quay.io/bjozsa-redhat/kubevirt-redfish:08ff5a7c"
  imagePullPolicy: "Always"
  # Service Type (TODO: will probably remove later)
  serviceType: "ClusterIP"
  
  # Route configuration (TODO: need to revisit non-OpenShift options later)
  routeEnabled: true
  # Comment out routeHost to let OpenShift auto-generate it (as per our previous fix)
  # routeHost: "kubevirt-redfish-development.apps.cluster.domain.com"
  
  # Resource requirements
  resources:
    requests:
      cpu: "100m"
      memory: "512Mi"
    limits:
      cpu: "500m"
      memory: "2Gi"
  
  # Chassis configuration
  chassis:
    - name: "development"
      namespace: "development"
      description: "jinkit KVM cluster with test VMs"
      serviceAccount: "kubevirt-redfish"
      vmSelector:
        redfish-enabled: "true"
  
  # Authentication configuration
  authentication:
    users:
      - username: "admin"
        passwordSecret: "redfish-admin-secret"
        chassis: ["development"]
  
  # TLS configuration (TODO: remove later)
  tls:
    enabled: false
  
  # Monitoring configuration (TODO: may refactor)
  monitoring:
    enabled: true
    serviceMonitor: true
    metricsPort: 8443
  
  # Virtual Media configuration
  virtualMedia:
    datavolume:
      storageSize: "3Gi"
      # TLS options (TODO: need to revisit non-OpenShift options later)
      allowInsecureTLS: true
      storageClass: "lvms-vg1"
      vmUpdateTimeout: "2m"
      isoDownloadTimeout: "30m"
      helperImage: "alpine:latest" 
```

Virtual media details:
- ISO image insertion and ejection
- DataVolume-based storage management
- Configurable timeouts and storage classes


### 5. Apply the RedfishServer Configuration

Apply the configuration:

```bash
kubectl apply -f redfishserver.yaml
```

### 6. Access the Redfish API

Get the Redfish server URL:

```bash
# For OpenShift
kubectl get route my-redfish-server -n my-namespace

# For Kubernetes
kubectl get svc my-redfish-server -n my-namespace
```

Test the API:

```bash
# Test root endpoint
curl -k -u admin:password https://your-redfish-url/redfish/v1/

# List systems
curl -k -u admin:password https://your-redfish-url/redfish/v1/Systems

# List Chassis
curl -k -u admin:password https://your-redfish-url/redfish/v1/Chassis

# List Systems associated with a Chassis (chassis=development)
curl -k -u admin:password https://your-redfish-url/redfish/v1/Chassis/development/Systems
```

## Configuration

### Chassis Configuration

```yaml
chassis:
  - name: "chassis-01"
    namespace: "development"
    description: "Description"
    serviceAccount: "service-account"
    vmSelector:
      redfish-enabled: "true"
      # Additional labels to select VMs
  - name: "chassis-02"
    namespace: "development"
    description: "production"
    serviceAccount: "service-account"
    vmSelector:
      redfish-enabled: "true"
      # Additional labels to select VMs
```

### Authentication

```yaml
authentication:
  users:
    - username: "admin"
      passwordSecret: "admin-secret"
      chassis: ["chassis-01"]
    - username: "user"
      passwordSecret: "user-secret"
      chassis: ["chassis-02"]
```

Create the password secrets:

```bash
# Create admin secret
kubectl create secret generic admin-secret \
  --from-literal=password=admin123 \
  -n development

# Create user secret
kubectl create secret generic user-secret \
  --from-literal=password=user123 \
  -n production
```

## Usage Examples

### Basic VM Management

```bash
# List all systems
curl -k -u admin:admin123 https://redfish-url/redfish/v1/Systems

# Get system details
curl -k -u admin:admin123 https://redfish-url/redfish/v1/Chassis/chassis-01/Systems/my-vm

# Power off a VM
curl -k -u admin:admin123 -X POST \
  -H "Content-Type: application/json" \
  -d '{"ResetType": "ForceOff"}' \
  https://redfish-url/redfish/v1/Chassis/chassis-01/Systems/my-vm/Actions/ComputerSystem.Reset

# Power on a VM
curl -k -u admin:admin123 -X POST \
  -H "Content-Type: application/json" \
  -d '{"ResetType": "On"}' \
  https://redfish-url/redfish/v1/Chassis/chassis-01/Systems/my-vm/Actions/ComputerSystem.Reset
```

## Troubleshooting

### Debugging

Enable debug logging:

```yaml
spec:
  # ... other fields ...
  env:
    - name: LOG_LEVEL
      value: "debug"
    - name: REDFISH_LOG_LEVEL
      value: "DEBUG"
```

### Health Checks

```bash
# Check operator health
kubectl get pods -n kubevirt-redfish-system

# Check RedfishServer health
kubectl get redfishserver -A

# Test Redfish API health
curl -k https://redfish-url/redfish/v1/
```

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/v1k0d3n/kubevirt-redfish-operator.git
cd kubevirt-redfish-operator

# Build the operator
make build

# Build and push container image
make build-push-version VERSION=v0.1.0
```

### Running Tests

```bash
# Run unit tests
make test

# Run integration tests
make test-integration
```

### Local Development

```bash
# Install CRDs
make install

# Deploy operator locally
make deploy

# Undeploy
make undeploy
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/v1k0d3n/kubevirt-redfish-operator/issues)
- **Documentation**: [Project Wiki](https://github.com/v1k0d3n/kubevirt-redfish-operator/wiki)
- **Discussions**: [GitHub Discussions](https://github.com/v1k0d3n/kubevirt-redfish-operator/discussions)

