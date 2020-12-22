pipeline {
    agent any
    stages {
        stage('Test this nonsense') {
            sh 'curl -s -H"Accept: application/vnd.github.v3+json" https://api.github.com/repos/virtual-vgo/vvgo/commits/${GIT_COMMIT}|jq -r .author.login>author'
            script {
                File file = new File("author")
                println(file.text)
            }
        }

        stage('Build & Test') {
            parallel {
                stage('Build Image') {
                    steps {
                        sh 'curl -s -H"Accept: application/vnd.github.v3+json" https://api.github.com/repos/virtual-vgo/vvgo/commits/${GIT_COMMIT}|jq -r .author.login>author'
                        script {
                            docker.withRegistry('https://ghcr.io', 'github_packages') {
                                def vvgoImage = docker.build("virtual-vgo/vvgo")
                                File file = new File("author")
                                vvgoImage.push('latest')
                                vvgoImage.push(GIT_COMMIT)
                                vvgoImage.push(BRANCH_NAME)
                                vvgoImage.push(file.text)
                            }
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

        stage('Deploy Staging') {
            steps { sh '/usr/bin/sudo /usr/bin/chef-solo -o vvgo::vvgo_staging' }
        }

        stage('Deploy Production') {
            when { branch 'master' }

            stages {
                stage('Deploy Container') {
                    steps { sh '/usr/bin/sudo /usr/bin/chef-solo -o vvgo::vvgo_prod' }
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
