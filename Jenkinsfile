// pipeline {
//     agent any
//
//     environment {
//         REGISTRY_ENDPOINT = credentials('docker-registry-endpoint')
//     }
//
//     stages {
//         stage('Update Components') {
//             steps {
//                 sh "docker pull golang:1.20.4-alpine3.18" // Update with current go image
//             }
//         }
//         stage('Build') {
//             steps {
//                 sh "docker build -t $REGISTRY_ENDPOINT/ystv/web-auth:$BUILD_ID ."
//             }
//         }
//         stage('Registry Upload') {
//             steps {
//                 sh "docker push $REGISTRY_ENDPOINT/ystv/web-auth:$BUILD_ID" // Uploaded to registry
//             }
//         }
//         stage('Deploy') {
//             stages {
//                 stage('Production') {
//                     when {
//                         expression { return env.TAG_NAME ==~ /v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)/ } // Checking if it is main semantic version release
//                     }
//                     environment {
//                         APP_ENV = credentials('wauth-prod-env')
//                     }
//                     steps {
//                         sshagent(credentials : ['deploy-web']) {
//                             script {
//                                 sh 'rsync -av $APP_ENV deploy@web:/data/webs/web-auth/env'
//                                 sh '''ssh -tt deploy@web << EOF
//                                     docker pull $REGISTRY_ENDPOINT/ystv/web-auth:$BUILD_ID
//                                     docker rm -f ystv-web-auth || true
//                                     docker run -d -p 1335:8080 --env-file /data/webs/web-auth/env -e WAUTH_VERSION=$TAG_NAME --name ystv-web-auth $REGISTRY_ENDPOINT/ystv/web-auth:$BUILD_ID
//                                     docker image prune -a -f --filter "label=site=auth"
//                                     exit
//                                 EOF'''
//                             }
//                         }
//                     }
//                 }
//                 stage('Development') {
//                     when {
//                         branch 'master'
//                         not {
//                             expression { return env.TAG_NAME ==~ /v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)/ }
//                         }
//                     }
//                     environment {
//                         APP_ENV = credentials('wauth-dev-env')
//                     }
//                     steps {
//                         sh "docker pull $REGISTRY_ENDPOINT/ystv/web-auth:$BUILD_ID" // Pulling image from registry
//                         script {
//                             try {
//                                 sh "docker rm -f ystv-web-auth" // Stop old container if it exists
//                             }
//                             catch (err) {
//                                 echo "Couldn't find container to stop"
//                                 echo err.getMessage()
//                             }
//                         }
//                         sh 'docker run -d -p 1335:8080 --env-file $APP_ENV -e WAUTH_VERSION=$TAG_NAME --name ystv-web-auth $REGISTRY_ENDPOINT/ystv/web-auth:$BUILD_ID' // Deploying site
//                         sh 'docker image prune -a -f --filter "label=site=auth"' // remove old image
//                     }
//                 }
//             }
//         }
//     }
//     post {
//         success {
//             echo 'Very cash-money'
//         }
//         failure {
//             echo 'That is not ideal, cheeky bugger'
//         }
//         always {
//             sh "docker image prune -f --filter label=site=auth --filter label=stage=builder" // Removing the local builder image
//         }
//     }
// }

String registryEndpoint = 'registry.comp.ystv.co.uk'

def image
String imageName = "ystv/web-auth:${env.BRANCH_NAME}-${env.BUILD_ID}"

pipeline {
  agent {
    label 'docker'
  }

  environment {
    DOCKER_BUILDKIT = '1'
  }

  stages {
    stage('Build image') {
      steps {
        script {
          docker.withRegistry('https://' + registryEndpoint, 'docker-registry') {
            image = docker.build(imageName)
          }
        }
      }
    }

    stage('Push image to registry') {
      steps {
        script {
          docker.withRegistry('https://' + registryEndpoint, 'docker-registry') {
            image.push()
            if (env.BRANCH_IS_PRIMARY) {
              image.push('latest')
            }
          }
        }
      }
    }

    stage('Deploy') {
      stages {
        stage('Development') {
          when {
            expression { env.BRANCH_IS_PRIMARY }
          }
          steps {
            build(job: 'Deploy Nomad Job', parameters: [
              string(name: 'JOB_FILE', value: 'web-auth-dev.nomad'),
              text(name: 'TAG_REPLACEMENTS', value: "${registryEndpoint}/${imageName}")
            ])
          }
        }

        stage('Production') {
          when {
            // Checking if it is semantic version release.
            expression { return env.TAG_NAME ==~ /v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)/ }
          }
          steps {
            build(job: 'Deploy Nomad Job', parameters: [
              string(name: 'JOB_FILE', value: 'web-auth-prod.nomad'),
              text(name: 'TAG_REPLACEMENTS', value: "${registryEndpoint}/${imageName}")
            ])
          }
        }
      }
    }
  }
}
