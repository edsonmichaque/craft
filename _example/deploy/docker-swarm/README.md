# Docker Swarm Deployment Guide

## Overview

This guide provides instructions for deploying craft using Docker Swarm in production, including setting up a CI/CD pipeline with GitHub Actions or GitLab CI.

## Prerequisites

- Docker and Docker Compose installed
- Access to a Docker Swarm cluster
- Docker image available in a registry

## Production Deployment

1. **Initialize Docker Swarm** (if not already initialized):
   ```bash
   docker swarm init
   ```

2. **Deploy the Stack**:
   - Use the following command to deploy the stack:
     ```bash
     docker stack deploy -c docker-compose.yml craft
     ```

3. **Monitor the Services**:
   - Use the following command to monitor the services:
     ```bash
     docker service ls
     ```

4. **Update the Stack**:
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