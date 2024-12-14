# craft Helm Chart

This Helm chart is used to deploy the craft application.

## Prerequisites

- Kubernetes 1.16+
- Helm 3.0+

## Installing the Chart

To install the chart with the release name `my-release`:

```bash
helm install my-release ./craft
```

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```bash
helm uninstall my-release
```

## Configuration

The following table lists the configurable parameters of the craft chart and their default values.

| Parameter          | Description                        | Default                |
| ------------------ | ---------------------------------- | ---------------------- |
| `image.repository` | Image repository                   | `your-registry/craft` |
| `image.tag`        | Image tag                          | `latest`               |
| `service.type`     | Kubernetes service type            | `ClusterIP`            |
| `service.port`     | Kubernetes service port            | `80`                   |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```bash
helm install my-release ./craft --set image.tag=1.0.0
```

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```bash
helm install my-release ./craft -f values.yaml
```

## Using in Local Environment

To use the Helm chart in a local Kubernetes environment (e.g., Minikube or Kind):

1. Start your local Kubernetes cluster.
2. Build and load the Docker image into your local cluster:
   ```bash
   eval $(minikube docker-env) # For Minikube
   docker build -t your-registry/craft:latest .
   ```
3. Install the Helm chart:
   ```bash
   helm install my-release ./craft
   ```
4. Access the application:
   ```bash
   minikube service my-release-craft
   ```

## Using in CI/CD

To use the Helm chart in a CI/CD pipeline:

1. Ensure your CI/CD environment has access to a Kubernetes cluster and Helm.
2. Build and push the Docker image to a registry accessible by your Kubernetes cluster.
3. Use the following steps in your CI/CD pipeline to deploy the application:

   ```yaml
   - name: Install Helm
     run: |
       curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash

   - name: Deploy to Kubernetes
     run: |
       helm upgrade --install my-release ./craft --set image.repository=your-registry/craft --set image.tag=${{ github.sha }}
   ```

Replace `your-registry/craft` with your actual Docker registry and image name.

## License

This Helm chart is licensed under the MIT License. See the LICENSE file for more details.
