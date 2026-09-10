// ─────────────────────────────────────────────────────────────────────────────
// s-erp-api — CI/CD
//
// Implements jenkins-nb/docs/ci-cd-implementation-plan.md.
//
//   Pull request → main     Pipeline A: full validation + AI review, no deploy
//   Push to main            Pipeline B: revalidate, build, scan, deploy, verify
//
// One file, two paths, because a multibranch job gives PR and branch builds the
// same checkout logic and credentials model. The separation the plan asks for
// (§31) is enforced by WHERE credentials are bound: no deploy or registry
// credential is ever bound inside a stage a PR build can reach.
//
// Commands come from .ci/config.yml, not from here (plan §4).
// ─────────────────────────────────────────────────────────────────────────────

@Library('nb-pipeline') _

def cfg = null
def results = []
def IMAGE_REF = ''

pipeline {
  agent { label 'docker' }

  options {
    timeout(time: 60, unit: 'MINUTES')
    timestamps()
    ansiColor('xterm')
    // Plan §32: a superseded PR commit should not keep burning an executor.
    // On master this is false — deploys must not be aborted mid-flight.
    //
    // `env.CHANGE_ID != null`, NOT `env.CHANGE_ID as Boolean`: on a branch build
    // CHANGE_ID is null, and Groovy's `null as Boolean` yields null rather than
    // false, which fails the build outright with
    //   Could not instantiate {abortPrevious=null} for DisableConcurrentBuildsJobProperty
    disableConcurrentBuilds(abortPrevious: env.CHANGE_ID != null)
    buildDiscarder(logRotator(numToKeepStr: '30', artifactNumToKeepStr: '10'))
    skipDefaultCheckout(false)
  }

  environment {
    APP_NAME  = 's-erp-api'
    REGISTRY  = 'ghcr.io'
    NAMESPACE = 'nibroos'
    // Immutable and traceable: build number sorts, SHA identifies. Plan §26.
    IMAGE_TAG = "${env.BUILD_NUMBER}-${env.GIT_COMMIT?.take(12)}"
  }

  stages {

    stage('Initialize') {
      steps {
        script {
          cfg = ciConfig()
          IMAGE_REF = "${REGISTRY}/${NAMESPACE}/${cfg.deployment.image}:${IMAGE_TAG}"

          currentBuild.displayName = "#${BUILD_NUMBER} ${env.GIT_COMMIT?.take(7)}"
          currentBuild.description = env.CHANGE_ID
              ? "PR #${env.CHANGE_ID} → ${env.CHANGE_TARGET}"
              : "branch ${env.BRANCH_NAME}"

          // Plan §6.1: read metadata from plugin-provided env vars, never from
          // anything a PR author can inject into a shell command.
          echo """
            |Repository : ${env.GIT_URL}
            |Branch     : ${env.BRANCH_NAME}
            |Commit     : ${env.GIT_COMMIT}
            |PR         : ${env.CHANGE_ID ?: '(none)'} ${env.CHANGE_ID ? "by ${env.CHANGE_AUTHOR}" : ''}
            |Target     : ${env.CHANGE_TARGET ?: '(n/a)'}
            |Image      : ${IMAGE_REF}
          """.stripMargin()

          if (env.CHANGE_ID) {
            githubStatus('CI / PR Quality Gate', 'pending', 'Validation running')
          }

          // Plan §5: only validate PRs aimed at the production branch.
          if (env.CHANGE_ID && env.CHANGE_TARGET != cfg.deployment.branch) {
            currentBuild.result = 'NOT_BUILT'
            error("PR targets '${env.CHANGE_TARGET}', not '${cfg.deployment.branch}' — not validated by this pipeline")
          }
        }
      }
    }

    stage('Install dependencies') {
      steps {
        // The Go commands write into these from inside service/, so they must
        // exist first — `go test | go-junit-report > ../reports/...` will not
        // create the directory for you.
        sh 'mkdir -p reports/junit coverage'
        script {
          def r = ciStage(name: 'Install', command: cfg.commands.install, timeout: 15)
          if (r.status == 'FAIL') {
            error 'Dependency installation failed — nothing downstream can be trusted'
          }
        }
      }
    }

    // Cheap, fast checks first: no reason to spend 15 minutes on Qodana for a
    // branch that does not compile. Plan §9.
    stage('Static checks') {
      parallel {
        stage('Lint') {
          steps { script { results << ciStage(name: 'Lint', command: cfg.commands.lint, timeout: 10) } }
        }
        stage('Type check') {
          steps { script { results << ciStage(name: 'Type check', command: cfg.commands.typecheck, timeout: 15) } }
        }
      }
    }

    stage('Tests & coverage') {
      steps {
        script {
          results << ciStage(name: 'Tests',
                             command: cfg.commands.coverage ?: cfg.commands.test,
                             junit: 'reports/junit/*.xml',
                             timeout: 20)
          // New-code coverage: judged against the branch this change targets,
          // so a PR is measured on the lines it touched.
          results << coverageReport(cobertura: 'coverage/cobertura-coverage.xml',
                                    baseBranch: env.CHANGE_TARGET ?: cfg.deployment.branch,
                                    minimum: cfg.coverage.minimum)
        }
      }
    }

    // Independent of one another and all slow — run them together.
    stage('Analysis') {
      parallel {
        stage('Semgrep')      { steps { script { results << semgrepScan(config: 'p/golang') } } }
        stage('Dependencies') { steps { script { results << depScan(runtime: cfg.runtime.type) } } }
        // Runs on PRs and on the production branch. The community branch
        // plugin gives Community Edition a branch model, so a PR is analysed as
        // a pull request and scored on its changed code only. Feature branches
        // outside a PR are skipped — they would just add noise to the project.
        stage('SonarQube') {
          when {
            allOf {
              expression { cfg.quality?.sonarqube }
              expression { env.CHANGE_ID || env.BRANCH_NAME == cfg.deployment.branch }
            }
          }
          steps { script { results << sonarScan(projectKey: env.APP_NAME) } }
        }
      }
    }

    // PR only. Plan §24: re-running the AI review on master adds cost without
    // adding a decision — the code was already reviewed on the PR.
    stage('AI review') {
      when { changeRequest() }
      steps { script { results << aiReview() } }
    }

    // One place decides pass/fail, from every collected result. Plan §19–20.
    stage('Quality gate') {
      steps { script { ciReport(config: cfg, results: results) } }
    }

    // ── Everything below is master-only. A PR build never reaches these
    //    stages, so it never binds registry or deploy credentials. Plan §31.
    stage('Build image') {
      when { allOf { branch "${cfg?.deployment?.branch ?: 'master'}"; expression { cfg.deployment.enabled } } }
      steps {
        sh '''
          set -eu
          docker buildx build \
            --file service/Dockerfile \
            --tag "${REGISTRY}/${NAMESPACE}/${APP_NAME}:${IMAGE_TAG}" \
            --label "org.opencontainers.image.revision=${GIT_COMMIT}" \
            --label "org.opencontainers.image.source=${GIT_URL}" \
            --label "org.opencontainers.image.created=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
            --cache-from "type=registry,ref=${REGISTRY}/${NAMESPACE}/${APP_NAME}:buildcache" \
            --cache-to   "type=registry,ref=${REGISTRY}/${NAMESPACE}/${APP_NAME}:buildcache,mode=max" \
            --load \
            service
        '''
      }
    }

    // Scan before push: a vulnerable image that is already in the registry can
    // be pulled by anything watching it. Plan §33.
    stage('Scan image') {
      when { allOf { branch "${cfg?.deployment?.branch ?: 'master'}"; expression { cfg.deployment.enabled } } }
      steps {
        script {
          def r = imageScan(image: IMAGE_REF)
          ciReport(config: cfg, results: [r], gate: true)
        }
      }
    }

    stage('Push image') {
      when { allOf { branch "${cfg?.deployment?.branch ?: 'master'}"; expression { cfg.deployment.enabled } } }
      steps {
        withCredentials([usernamePassword(credentialsId: 'registry-credentials',
                                          usernameVariable: 'REG_USER',
                                          passwordVariable: 'REG_TOKEN')]) {
          sh '''
            set -eu
            echo "$REG_TOKEN" | docker login "$REGISTRY" -u "$REG_USER" --password-stdin
            docker push "${REGISTRY}/${NAMESPACE}/${APP_NAME}:${IMAGE_TAG}"
            docker logout "$REGISTRY"
          '''
        }
      }
    }

    stage('Deploy — this host') {
      when { allOf { branch "${cfg?.deployment?.branch ?: 'master'}"; expression { cfg.deployment.enabled } } }
      options { lock(resource: 's-erp-api-local') }   // plan §32: never concurrent
      steps {
        script {
          dockerDeploy(target: 'local', app: cfg.deployment.app, image: IMAGE_REF,
                       health: "http://localhost:${env.APP_PORT ?: '3002'}/healthz")
        }
      }
    }

    stage('Approve production') {
      when { allOf { branch "${cfg?.deployment?.branch ?: 'master'}"; expression { cfg.deployment.enabled } } }
      options { timeout(time: 2, unit: 'HOURS') }
      steps {
        script {
          input message: "Deploy ${IMAGE_REF} to the production host?",
                ok: 'Deploy',
                submitterParameter: 'APPROVER'
        }
      }
    }

    stage('Deploy — production host') {
      when { allOf { branch "${cfg?.deployment?.branch ?: 'master'}"; expression { cfg.deployment.enabled } } }
      options { lock(resource: 's-erp-api-production') }
      steps {
        script {
          // deploy.sh health-gates and rolls back automatically on failure.
          dockerDeploy(target: 'remote', app: cfg.deployment.app, image: IMAGE_REF,
                       health: "http://${env.PROD_HOST ?: 'localhost'}:${env.APP_PORT ?: '3002'}/healthz")
          echo "Deployed ${IMAGE_REF} · commit ${env.GIT_COMMIT} · approved by ${env.APPROVER ?: 'n/a'}"
        }
      }
    }
  }

  post {
    always {
      // notFailBuild on both: a post-build housekeeping problem must not be the
      // reason a green pipeline reports red.
      archiveArtifacts artifacts: 'reports/**', allowEmptyArchive: true
      cleanWs(deleteDirs: true, notFailBuild: true, disableDeferredWipeout: true,
              patterns: [[pattern: '.qodana/**', type: 'EXCLUDE']])
    }
    failure {
      script {
        if (env.CHANGE_ID) {
          githubStatus('CI / PR Quality Gate', 'failure', 'Pipeline failed — see the build log')
        } else {
          // env.APP_NAME, not APP_NAME: inside a post block the environment is
          // not in the script binding, and the bare name throws
          // MissingPropertyException while handling another failure.
          echo "${env.BRANCH_NAME} build failed. If a deploy ran, deploy.sh has " +
               "already rolled back. Verify with: ops/rollback → app=${env.APP_NAME}"
        }
      }
    }
  }
}
