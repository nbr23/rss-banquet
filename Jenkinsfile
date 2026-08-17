pipeline {
    agent any
    options {
        disableConcurrentBuilds()
    }
    stages {
        stage('Checkout'){
            steps {
                checkout scm
            }
        }
        stage('Prep buildx') {
            when { branch 'master' }
            steps {
                script {
                    env.BUILDX_BUILDER = getBuildxBuilder();
                }
            }
        }
        stage('Test') {
            when {
                branch 'master'
                not { changelog '^skip-tests.*' }
            }
            steps {
                sh """
                    docker buildx build --pull --builder \$BUILDX_BUILDER --build-arg BANQUET_GLOBAL_USER_AGENT='Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:143.0) Gecko/20100101 Firefox/143.0' --target test -t rss-banquet-test .
                    """
            }
        }
        stage('Dockerhub login') {
            when { branch 'master' }
            steps {
                withCredentials([usernamePassword(credentialsId: 'dockerhub', usernameVariable: 'DOCKERHUB_CREDENTIALS_USR', passwordVariable: 'DOCKERHUB_CREDENTIALS_PSW')]) {
                    sh 'docker login -u $DOCKERHUB_CREDENTIALS_USR -p "$DOCKERHUB_CREDENTIALS_PSW"'
                }
            }
        }
        stage('Build Base Docker Image') {
            when { branch 'master' }
            steps {
                sh """
                    docker buildx build --pull --builder \$BUILDX_BUILDER  --target base --platform linux/arm64,linux/amd64 -t nbr23/rss-banquet:latest -t nbr23/rss-banquet:`git rev-parse --short HEAD` --push .
                    """
            }
        }
        stage('Build Server Docker Image') {
            when { branch 'master' }
            steps {
                sh """
                    docker buildx build --pull --builder \$BUILDX_BUILDER  --target server --platform linux/arm64,linux/amd64 -t nbr23/rss-banquet:server-latest -t nbr23/rss-banquet:server-`git rev-parse --short HEAD` --push .
                    """
            }
        }
        stage('Build Nginx Server Docker Image') {
            when { branch 'master' }
            steps {
                sh """
                    docker buildx build --pull --builder \$BUILDX_BUILDER  --target nginx --platform linux/arm64,linux/amd64 -t nbr23/rss-banquet:server-nginx-latest -t nbr23/rss-banquet:server-nginx-`git rev-parse --short HEAD` --push .
                    """
            }
        }
        stage('Sync github repos') {
            when { branch 'master' }
            steps {
                ghSync()
            }
        }
    }
}
