pipeline {
    agent any

    environment {
        APP_NAME   = 'pos-backend'
        GO_VERSION = '1.26'
        // Docker image name — adjust to your registry
        IMAGE_NAME = "your-registry/${APP_NAME}"
    }

    stages {
        // ── Stage 1: Checkout ─────────────────────────────────────────────────
        stage('Checkout') {
            steps {
                checkout scm
                echo "Branch: ${env.BRANCH_NAME} | Commit: ${env.GIT_COMMIT[0..7]}"
            }
        }

        // ── Stage 2: Setup Go ─────────────────────────────────────────────────
        stage('Setup') {
            steps {
                sh 'go version'
                sh 'go mod download'
                sh 'go mod verify'
            }
        }

        // ── Stage 3: Lint ─────────────────────────────────────────────────────
        stage('Lint') {
            steps {
                sh '''
                    if ! command -v golangci-lint &> /dev/null; then
                        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
                    fi
                    golangci-lint run ./...
                '''
            }
        }

        // ── Stage 4: Test ─────────────────────────────────────────────────────
        stage('Test') {
            steps {
                sh 'go test -v -race -coverprofile=coverage.out ./...'
            }
            post {
                always {
                    // Publish coverage report if JUnit plugin is installed
                    sh 'go tool cover -func=coverage.out'
                }
            }
        }

        // ── Stage 5: Build ────────────────────────────────────────────────────
        stage('Build') {
            steps {
                sh '''
                    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
                    go build -ldflags="-s -w" -o bin/${APP_NAME} ./cmd/api/main.go
                '''
            }
        }

        // ── Stage 6: Docker Build & Push (only on main) ───────────────────────
        stage('Docker') {
            when {
                branch 'main'
            }
            steps {
                script {
                    def imageTag = "${IMAGE_NAME}:${env.GIT_COMMIT[0..7]}"
                    def latestTag = "${IMAGE_NAME}:latest"

                    sh "docker build -t ${imageTag} -t ${latestTag} ."
                    // Uncomment when Docker registry credentials are configured:
                    // docker.withRegistry('https://your-registry', 'registry-credentials') {
                    //     sh "docker push ${imageTag}"
                    //     sh "docker push ${latestTag}"
                    // }
                }
            }
        }

        // ── Stage 7: Deploy (only on main) ───────────────────────────────────
        stage('Deploy') {
            when {
                branch 'main'
            }
            steps {
                // TODO: add deployment steps (e.g., kubectl apply, docker-compose pull+up, etc.)
                echo 'Deploy step — to be configured based on infrastructure'
            }
        }
    }

    post {
        always {
            cleanWs()
        }
        success {
            echo "Pipeline completed successfully for ${env.BRANCH_NAME}"
        }
        failure {
            echo "Pipeline failed for ${env.BRANCH_NAME} — check logs above"
            // TODO: add notification (Slack, email, etc.)
        }
    }
}
