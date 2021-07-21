pipeline {
    agent any

    stages {
        stage('Update Components') {
            steps {
                sh "docker pull golang:1.16-alpine" // Update with current go image
            }
        }
        stage('Build') {
            steps {
                sh "docker build -t localhost:5000/ystv/web-auth:$BUILD_ID ."
            }
        }
        stage('Registry Upload') {
            steps {
                sh "docker push localhost:5000/ystv/web-auth:$BUILD_ID" // Uploaded to registry
            }
        }
        stage('Deploy') {
            stages {
                stage('Production') {
                    when {
                        branch 'master'
                        tag pattern: "^v(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)", comparator: "REGEXP" // Checking if it is main semantic version release
                    }
                    environment {
                        APP_ENV = credentials('wauth-prod-env')
                    }
                    steps {
                        sshagent(credentials : ['deploy-web']) {
                            script {
                                sh 'rsync -av $APP_ENV deploy@web:/data/webs/web-auth/env'
                                sh '''ssh -tt deploy@web << EOF
                                    docker pull localhost:5000/ystv/web-auth:$BUILD_ID
                                    docker rm -f ystv-web-auth || true
                                    docker run -d -p 1335:8080 --env-file /data/webs/web-auth/env --name ystv-web-auth localhost:5000/ystv/web-auth:$BUILD_ID
                                    docker image prune -a -f --filter "label=site=auth"
                                EOF'''
                            }
                        }
                    }
                }
                stage('Development') {
                    when {
                        branch 'master'
                        not {
                            tag pattern: "^^v(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)", comparator: "REGEXP"
                        }
                    }
                    environment {
                        APP_ENV = credentials('wauth-dev-env')
                    }
                    steps {
                        sh "docker pull localhost:5000/ystv/web-auth:$BUILD_ID" // Pulling image from registry
                        script {
                            try {
                                sh "docker rm -f ystv-web-auth" // Stop old container if it exists
                            }
                            catch (err) {
                                echo "Couldn't find container to stop"
                                echo err.getMessage()
                            }
                        }
                        sh 'docker run -d -p 1335:8080 --env-file $APP_ENV --name ystv-web-auth localhost:5000/ystv/web-auth:$BUILD_ID' // Deploying site
                        sh 'docker image prune -a -f --filter "label=site=auth"' // remove old image
                    }
                }
            }   
        }
    }
    post {
        success {
            echo 'Very cash-money'
        }
        failure {
            echo 'That is not ideal, cheeky bugger'
        }
        always {
            sh "docker image prune -f --filter label=site=auth --filter label=stage=builder" // Removing the local builder image
        }
    }
}
