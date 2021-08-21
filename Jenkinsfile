pipeline {
    agent any

    environment {
        SSH_CREDS = credentials('jenkins_ssh_key')
        DEPLOY_TARGET = 'jenkins@vvgo-2.infra.vvgo.org'
    }

    stages {
        stage('Unit Test') {
            steps {
                script {
                    docker.withRegistry('https://ghcr.io', 'github_packages') {
                        docker
                            .build("virtual-vgo/vvgo-builder:${GIT_COMMIT}", "--target builder -f Dockerfile .")
                            .inside("-u root --network test-network -e REDIS_ADDRESS=redis-testing:6379 -e MINIO_ENDPOINT=minio-testing:9000") {
                                sh 'go generate ./...'
                                sh 'go test -v ./...'
                        }
                    }
                }
            }
        }

        stage('Build Image') {
            steps {
                script {
                    def author = sh(
                        script: '''
                            curl -s  -H "Accept: application/vnd.github.v3+json"  \
                            https://api.github.com/repos/virtual-vgo/vvgo/commits/${GIT_COMMIT}|jq -r .author.login
                        ''', returnStdout: true)
                    docker.withRegistry('https://ghcr.io', 'github_packages') {
                        def vvgoImage = docker.build("virtual-vgo/vvgo")
                        vvgoImage.push('latest')
                        vvgoImage.push(GIT_COMMIT)
                        vvgoImage.push(BRANCH_NAME)
                        if (author != "") { vvgoImage.push(author) }
                    }
                }
            }
        }

        stage('Deploy Staging') {
            when { not { branch 'master' } }
            steps { sh 'ssh -i ${SSH_CREDS} ${DEPLOY_TARGET} sudo /usr/local/bin/chef-solo -o vvgo::staging' }
        }

        stage('Deploy Production') {
            when { branch 'master' }
            steps { sh 'ssh -i ${SSH_CREDS} ${DEPLOY_TARGET} sudo /usr/local/bin/chef-solo -o vvgo::prod' }
            post {
                success {
                    withCredentials(bindings: [string(credentialsId: 'web_and_coding_team_webhook', variable: 'WEBHOOK_URL')]) {
                        discordSend(link: env.BUILD_URL,
                                result: currentBuild.currentResult,
                                title: "vvgo build ${BUILD_NUMBER} deployed",
                                webhookURL: "${WEBHOOK_URL}")
                    }
                }

                unsuccessful {
                    withCredentials(bindings: [string(credentialsId: 'web_and_coding_team_webhook', variable: 'WEBHOOK_URL')]) {
                        discordSend(link: env.BUILD_URL,
                                result: currentBuild.currentResult,
                                title: "vvgo build ${BUILD_NUMBER} has failures",
                                webhookURL: '${WEBHOOK_URL}')
                    }
                }
            }
        }

        stage('Purge Cloudflare Cache') {
            when { branch 'master' }
            steps {
                withCredentials(bindings: [string(credentialsId: 'cloudflare_purge_key', variable: 'API_KEY')]) {
                    httpRequest(httpMode: 'POST',
                            contentType: 'APPLICATION_JSON',
                            customHeaders: [[name: 'Authorization', value: "Bearer ${API_KEY}"]],
                            requestBody: '{"purge_everything":true}',
                            url: 'https://api.cloudflare.com/client/v4/zones/e3cfa4eadcdea773633d52a52cb6203f/purge_cache')
                }
            }
        }
    }
}
