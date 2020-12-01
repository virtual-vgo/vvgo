pipeline {
    agent any
    stages {
        stage('Build') {
            steps {
                sh 'docker build . -t vvgo:latest -t vvgo:${BRANCH_NAME}'
            }
        }
        stage('Deploy') {
            when {
                branch 'master'
            }
            stages {
                stage('Deploy container')
                steps {
                    sh 'docker rm -f vvgo-prod || true'
                    sh '''
                        docker run -d --name vvgo-prod \
                            --env GOOGLE_APPLICATION_CREDENTIALS=/etc/vvgo/google_api_credentials.json \
                            --volume /etc/vvgo:/etc/vvgo \
                            --publish 8080:8080 \
                            --network prod-network \
                            vvgo:master
                    '''
                }
            }
            stage('Purge cache') {
                steps {
                    withCredentials(bindings: [string(credentialsId: 'cloudflare_purge_key', variable: 'API_KEY')]) {
                        httpRequest(httpMode: 'POST',
                                contentType: 'application/json',
                                customHeaders: [[name: 'Authorization', value: "Bearer ${API_KEY}"]],
                                requestBody: '{"purge_everything":true}',
                                url: 'https://api.cloudflare.com/client/v4/zones/e3cfa4eadcdea773633d52a52cb6203f/purge_cache')
                    }

                }
            }
        }
        post {
            success {
                discordSend(link: env.BUILD_URL, result: currentBuild.currentResult, title: "vvgo build ${BUILD_NUMBER} deployed", webhookURL: 'https://discordapp.com/api/webhooks/759951149285113857/Zx7awCsOqvph30i-96i2S19v9Ax6Yc0LtXAer9k9C2ZEGJOq6tClgoY05aEkgxkE0X7y')
            }
            failure {
                discordSend(link: env.BUILD_URL, result: currentBuild.currentResult, title: "vvgo build ${BUILD_NUMBER} has failures", webhookURL: 'https://discordapp.com/api/webhooks/759951149285113857/Zx7awCsOqvph30i-96i2S19v9Ax6Yc0LtXAer9k9C2ZEGJOq6tClgoY05aEkgxkE0X7y')
            }
        }
    }
}
