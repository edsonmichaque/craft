package main

import (
    "dagger.io/dagger"
    "universe.dagger.io/docker"
    "universe.dagger.io/bash"
)

// Project configuration
#Config: {
    projectName: string | *"craft"
}

config: #Config

// Base container with CI scripts
#BaseContainer: {
    docker.#Build & {
        steps: [
            docker.#Pull & {
                source: "golang:1.21"
            },
            docker.#Copy & {
                contents: client.filesystem."./".read.contents
                dest: "/workspace"
            },
            // Ensure scripts are executable
            bash.#Run & {
                script: contents: """
                    chmod -R +x /workspace/scripts/ci
                    """
            },
        ]
    }
}

dagger.#Plan & {
    client: {
        filesystem: {
            "./": read: {
                contents: dagger.#FS
                exclude: [
                    "bin",
                    "dist",
                    "coverage",
                    ".git",
                ]
            }
        }
        env: {
            CI:               string | *"true"
            CI_COMMIT_SHA:    string | *""
            CI_COMMIT_TAG:    string | *""
            CI_COMMIT_BRANCH: string | *""
            LOG_LEVEL:        string | *"INFO"
        }
    }

    actions: {
        // Build task using build.sh
        build: docker.#Run & {
            input: #BaseContainer
            workdir: "/workspace"
            command: {
                name: "./scripts/ci/tasks/build.sh"
            }
        }

        // Test task using test.sh
        test: docker.#Run & {
            input: #BaseContainer
            workdir: "/workspace"
            command: {
                name: "./scripts/ci/tasks/test.sh"
            }
        }

        // Lint task using lint.sh
        lint: docker.#Run & {
            input: #BaseContainer
            workdir: "/workspace"
            command: {
                name: "./scripts/ci/tasks/lint.sh"
            }
        }

        // Docker task using docker.sh
        docker: {
            build: docker.#Run & {
                input: #BaseContainer
                workdir: "/workspace"
                command: {
                    name: "./scripts/ci/tasks/docker.sh"
                    args: ["build"]
                }
            }

            push: docker.#Run & {
                input: #BaseContainer
                workdir: "/workspace"
                command: {
                    name: "./scripts/ci/tasks/docker.sh"
                    args: ["push"]
                }
            }
        }
        }

        // Release task using release.sh
        release: docker.#Run & {
            input: #BaseContainer
            workdir: "/workspace"
            command: {
                name: "./scripts/ci/tasks/release.sh"
            }
        }

        // Proto task using proto.sh
        proto: docker.#Run & {
            input: #BaseContainer
            workdir: "/workspace"
            command: {
                name: "./scripts/ci/tasks/proto.sh"
            }
        }

        // Dependencies task using dependencies.sh
        dependencies: docker.#Run & {
            input: #BaseContainer
            workdir: "/workspace"
            command: {
                name: "./scripts/ci/tasks/dependencies.sh"
            }
        }

        // Package task using package.sh
        package: docker.#Run & {
            input: #BaseContainer
            workdir: "/workspace"
            command: {
                name: "./scripts/ci/tasks/package.sh"
            }
        }

        // Utils
        utils: {
            // Health check using health-check.sh
            healthCheck: docker.#Run & {
                input: #BaseContainer
                workdir: "/workspace"
                command: {
                    name: "./scripts/ci/utils/health-check.sh"
                }
            }

            // Cleanup using cleanup.sh
            cleanup: docker.#Run & {
                input: #BaseContainer
                workdir: "/workspace"
                command: {
                    name: "./scripts/ci/utils/cleanup.sh"
                }
            }

            // Setup dev using setup-dev.sh
            setupDev: docker.#Run & {
                input: #BaseContainer
                workdir: "/workspace"
                command: {
                    name: "./scripts/ci/utils/setup-dev.sh"
                }
            }
        }

        // CI Pipeline
        ci: {
            pipeline: dagger.#Pipeline & {
                steps: [
                    utils.cleanup,
                    dependencies,
                    lint,
                    test,
                    build,
                    docker.build,
                    if client.env.CI_COMMIT_TAG != "" {
                        package
                    },
                ]
            }
        }
    }
}