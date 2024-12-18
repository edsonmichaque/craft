# Docker Swarm Deployment Guide

## Overview

This guide provides detailed instructions for installing and configuring Docker Swarm, and deploying craft using Docker Swarm in production. It also includes setting up a CI/CD pipeline with GitHub Actions or GitLab CI.

## Prerequisites

- **Operating System**: Ensure you are using a Linux-based OS (e.g., Ubuntu, CentOS).
- **Docker and Docker Compose**: Installed on all nodes.
- **Access to a Docker Swarm cluster**: At least one manager and one worker node.
- **Docker image available in a registry**: Ensure your application image is built and pushed to a Docker registry.

## Installing Docker

1. **Update your package index**:
   ```bash
   sudo apt-get update
   ```

2. **Install Docker**:
   ```bash
   sudo apt-get install -y docker.io
   ```

3. **Start Docker and enable it to start at boot**:
   ```bash
   sudo systemctl start docker
   sudo systemctl enable docker
   ```

4. **Verify Docker installation**:
   ```bash
   docker --version
   ```

## Configuring Docker Swarm

1. **Initialize Docker Swarm** (on the manager node):
   ```bash
   docker swarm init
   ```

   - If you have multiple nodes, note the `docker swarm join` command output. You'll use this to add worker nodes.

2. **Join Worker Nodes**:
   - Run the `docker swarm join` command on each worker node:
     ```bash
     docker swarm join --token <token> <manager-ip>:2377
     ```

3. **Verify Nodes**:
   - On the manager node, list all nodes to ensure they are part of the swarm:
     ```bash
     docker node ls
     ```

## Production Deployment

1. **Deploy the Stack**:
   - Use the following command to deploy the stack:
     ```bash
     docker stack deploy -c docker-compose.yml craft
     ```

2. **Monitor the Services**:
   - Use the following command to monitor the services:
     ```bash
     docker service ls
     ```

3. **Update the Stack**:
   - To update the stack with new configurations or images:
     ```bash
     docker stack deploy -c docker-compose.yml craft
     ```

## Setting Up CI/CD for Production

### Using GitHub Actions

1. **Create a Workflow File**:
   - Create a `.github/workflows/deploy.yml` file in your repository.

2. **Configure the Workflow**:
   - Add the following steps to build, push, and deploy your Docker image:
     ```yaml
     name: CI/CD Pipeline

     on:
       push:
         branches: [ main ]

     jobs:
       build:
         runs-on: ubuntu-latest
         steps:
           - uses: actions/checkout@v4

           - name: Set up Docker Buildx
             uses: docker/setup-buildx-action@v3

           - name: Log in to Docker Hub
             uses: docker/login-action@v3
             with:
               username: ${{ secrets.DOCKER_USERNAME }}
               password: ${{ secrets.DOCKER_PASSWORD }}

           - name: Build and push Docker image
             uses: docker/build-push-action@v4
             with:
               context: .
               push: true
               tags: ${{ secrets.DOCKER_REGISTRY }}/craft:latest

       deploy:
         runs-on: ubuntu-latest
         needs: build
         steps:
           - name: SSH to Swarm Manager
             uses: appleboy/ssh-action@v0.1.3
             with:
               host: ${{ secrets.SWARM_MANAGER_HOST }}
               username: ${{ secrets.SSH_USERNAME }}
               key: ${{ secrets.SSH_PRIVATE_KEY }}
               script: |
                 docker stack deploy -c /path/to/docker-compose.yml craft
     ```

3. **Configure Secrets**:
   - Add the following secrets to your GitHub repository:
     - `DOCKER_USERNAME`: Your Docker Hub username
     - `DOCKER_PASSWORD`: Your Docker Hub password
     - `DOCKER_REGISTRY`: Your Docker registry URL
     - `SWARM_MANAGER_HOST`: The IP address of your Docker Swarm manager node
     - `SSH_USERNAME`: The SSH username for accessing the Swarm manager
     - `SSH_PRIVATE_KEY`: The SSH private key for accessing the Swarm manager

### Using GitLab CI

1. **Create a `.gitlab-ci.yml` File**:
   - Add the following configuration to build, push, and deploy your Docker image:
     ```yaml
     stages:
       - build
       - deploy

     variables:
       DOCKER_DRIVER: overlay2

     build:
       stage: build
       script:
         - docker build -t $CI_REGISTRY_IMAGE:latest .
         - docker push $CI_REGISTRY_IMAGE:latest

     deploy:
       stage: deploy
       script:
         - ssh $SSH_USERNAME@$SWARM_MANAGER_HOST "docker stack deploy -c /path/to/docker-compose.yml craft"
       only:
         - main
     ```

2. **Configure Variables**:
   - Set the following variables in your GitLab CI/CD settings:
     - `CI_REGISTRY_IMAGE`: Your Docker registry image path
     - `SSH_USERNAME`: The SSH username for accessing the Swarm manager
     - `SWARM_MANAGER_HOST`: The IP address of your Docker Swarm manager node
     - `SSH_PRIVATE_KEY`: The SSH private key for accessing the Swarm manager

## Additional Commands

- **Remove the Stack**:
  ```bash
  docker stack rm craft
  ```

- **List Nodes**:
  ```bash
  docker node ls
  ```

- **Inspect a Service**:
  ```bash
  docker service inspect --pretty craft_craft
  ```

## Notes

- Ensure that the Docker image '${{ .DockerImage }}' is available in your Docker registry.
- Adjust the number of replicas and other configurations as needed for your environment.