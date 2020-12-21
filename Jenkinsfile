def dockerRegistry = docker.withRegistry('https://docker.pkg.github.com', 'github_packages')

pipeline {
    agent any
    stages {
        stage('Build & Test') {
            parallel {
                stage('Build Image') {
                    steps {
                        script {
                            def vvgoImage = docker.build("docker.pkg.github.com/virtual-vgo/vvgo/vvgo")
                            vvgoImage.push('latest')
                            vvgoImage.push(GIT_COMMIT)
                            vvgoImage.push(BRANCH_NAME)
                        }
                    }
                }

                stage('Run Unit Tests') {
                    agent {
                        docker {
                            image 'golang:1.14'
                            args  "-v go-pkg-cache:/go/pkg -v go-build-cache:/.cache/go-build --network test-network"
                        }
                    }
                    environment {
                        REDIS_ADDRESS  = 'redis-testing:6379'
                        MINIO_ENDPOINT = 'minio-testing:9000'
                    }
                    steps {
                        sh 'go generate ./...'
                        sh 'go get -u github.com/jstemmer/go-junit-report'
                        sh 'go test -v -race ./... 2>&1 | go-junit-report > report.xml'
                        junit 'report.xml'
                    }
                }
            }
        }

        stage('Deploy') {
            when {
                branch 'master'
            }

            stages {
                stage('Deploy Container') {
                    steps {
                        sh '/usr/bin/sudo /usr/bin/chef-solo -o vvgo::vvgo_prod'
                    }
                }

                stage('Purge Cloudflare Cache') {
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

            post {
                success {
                    withCredentials(bindings: [string(credentialsId: 'web_and_coding_team_webhook', variable: 'WEBHOOK_URL')]) {
                        discordSend(link: env.BUILD_URL,
                                result: currentBuild.currentResult,
                                title: "vvgo build ${BUILD_NUMBER} deployed",
                                webhookURL: "${WEBHOOK_URL}")
                    }
                }

                failure {
                    withCredentials(bindings: [string(credentialsId: 'web_and_coding_team_webhook', variable: 'WEBHOOK_URL')]) {
                        discordSend(link: env.BUILD_URL,
                                result: currentBuild.currentResult,
                                title: "vvgo build ${BUILD_NUMBER} has failures",
                                webhookURL: "${WEBHOOK_URL}")
                    }
                }
            }
        }
    }
}
