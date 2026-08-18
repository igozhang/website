# k8s_cicd

```
将 Jenkins、私有 Harbor 镜像仓库和 Rancher API 结合，可以构建一个安全、自动化的 Kubernetes 应用发布流程。这个流程的核心在于：Jenkins 负责构建和推送镜像，而 Rancher API 则负责安全地触发 Kubernetes 集群的更新，避免了直接操作集群的权限和安全风险。
```

下面是一个典型的发布流程举例说明：

### 🚀 整体流程概览

1. **代码提交**: 开发人员将代码推送到 Git 仓库（如 GitLab）。
2. **触发构建**: Git 仓库的 Webhook 自动触发 Jenkins 流水线任务。
3. **构建与推送**: Jenkins 拉取代码，构建 Docker 镜像，并推送到私有 Harbor 仓库。
4. **API 发布**: Jenkins 调用 Rancher API，通知其更新指定工作负载（Workload）的镜像版本。
5. **自动部署**: Rancher 接收到指令后，自动从 Harbor 拉取新镜像并更新 Kubernetes 中的应用。

---

### 🛠️ 前期准备工作

在开始之前，需要完成以下配置：

1. **在 Harbor 中创建项目和用户**
   - 创建一个私有项目（例如 `my-project`）。
   - 创建一个用户（例如 `jenkins-bot`），并将其添加到项目中，赋予其“开发者”或更高角色，以确保有推送镜像的权限。
2. **在 Rancher 中创建 API Key**
   - 登录 Rancher UI，点击用户头像 -> `API & Keys`。
   - 点击 `添加密钥`，填写描述（如 `jenkins-api-key`），选择“永不过期”。
   - **重要**：创建成功后，Rancher 会显示 `Access Key`（用户名）和 `Secret Key`（密码），请务必妥善保存，因为它们只会显示一次。
3. **在 Jenkins 中配置凭据**
   - 进入 Jenkins 的 `系统管理` -> `凭据`。
   - 添加 Harbor 的用户名和密码，用于登录和推送镜像。
   - 安装 `Redeploy Rancher2.x Workload` 插件，并添加 Rancher API Key 凭据（类型选择 `Rancher2.x API Keys`），填入上一步获取的 `Access Key` 和 `Secret Key`。

---

### 📜 Jenkins 流水线 (Jenkinsfile) 示例

以下是一个 `Jenkinsfile` 的示例，它清晰地展示了整个自动化流程。

```
pipeline {
    agent any

    environment {
        // 定义全局变量，方便后续使用
        HARBOR_ADDR = 'harbor.example.com'
        HARBOR_PROJECT = 'my-project'
        IMAGE_NAME = 'my-app'
        RANCHER_CLUSTER_ID = 'c-xxxxx' // Rancher 集群ID
        RANCHER_PROJECT_ID = 'p-xxxxx' // Rancher 项目ID
        RANCHER_NAMESPACE = 'my-namespace' // K8s 命名空间
        RANCHER_WORKLOAD_NAME = 'my-app-deployment' // 工作负载名称
    }

    stages {
        // 阶段 1: 从 Git 仓库拉取源代码
        stage('拉取代码') {
            steps {
                // 这里假设已经配置了 Git 凭据
                git url: 'http://gitlab.example.com/my-group/my-app.git', branch: 'main'
            }
        }

        // 阶段 2: 构建 Docker 镜像
        stage('构建镜像') {
            steps {
                script {
                    // 使用 Git 提交哈希作为镜像标签，确保唯一性
                    def buildTag = sh(returnStdout: true, script: 'git rev-parse --short HEAD').trim()
                    env.BUILD_TAG = buildTag
                    sh "docker build -t ${IMAGE_NAME}:${BUILD_TAG} ."
                }
            }
        }

        // 阶段 3: 登录 Harbor 并推送镜像
        stage('推送镜像到 Harbor') {
            steps {
                // 使用 withCredentials 安全地注入 Harbor 凭据
                withCredentials([usernamePassword(credentialsId: 'harbor-credentials-id', usernameVariable: 'HARBOR_USER', passwordVariable: 'HARBOR_PASS')]) {
                    sh '''
                        # 登录 Harbor
                        echo ${HARBOR_PASS} | docker login ${HARBOR_ADDR} -u ${HARBOR_USER} --password-stdin
                        # 为镜像打上 Harbor 地址的标签
                        docker tag ${IMAGE_NAME}:${BUILD_TAG} ${HARBOR_ADDR}/${HARBOR_PROJECT}/${IMAGE_NAME}:${BUILD_TAG}
                        # 推送镜像
                        docker push ${HARBOR_ADDR}/${HARBOR_PROJECT}/${IMAGE_NAME}:${BUILD_TAG}
                    '''
                }
            }
        }

        // 阶段 4: 调用 Rancher API 更新工作负载
        stage('通过 Rancher API 发布') {
            steps {
                // 使用 Rancher 插件更新工作负载的镜像
                redeployRancher2Workload(
                    rancherUrl: 'https://rancher.example.com', // Rancher 访问地址
                    credentialId: 'rancher-api-key-id',       // Jenkins 中配置的 Rancher API 凭据 ID
                    clusterId: "${RANCHER_CLUSTER_ID}",
                    projectId: "${RANCHER_PROJECT_ID}",
                    namespace: "${RANCHER_NAMESPACE}",
                    workloadName: "${RANCHER_WORKLOAD_NAME}",
                    // 指定新的镜像地址和标签
                    containerImage: "${HARBOR_ADDR}/${HARBOR_PROJECT}/${IMAGE_NAME}:${BUILD_TAG}"
                )
            }
        }
    }
}
```

### ✨ 流程关键点解析

- **安全性**：
  - **Harbor 凭据**：通过 Jenkins 的 `withCredentials` 插件来管理，避免在脚本中硬编码密码。
  - **Rancher API**：使用 Rancher API 发布，Jenkins 无需持有 K8s 集群的 `kubeconfig` 或通过 SSH 登录到 Master 节点，大大降低了安全风险。
- **镜像版本控制**：使用 Git 提交的短哈希（`git rev-parse --short HEAD`）作为 Docker 镜像的标签，可以精确地将代码版本与镜像版本关联起来，便于追溯和管理。
- **自动化**：整个流程从代码提交到应用更新完全自动化，无需人工干预，提高了发布效率和一致性。
- **K8s 镜像拉取**：在 Rancher 中部署工作负载时，需要提前配置好 `imagePullSecrets`，让 Kubernetes 能够使用 Harbor 的用户凭据来拉取私有镜像。这通常在 Rancher UI 中为命名空间配置 Registry 凭据即可，Rancher 会自动处理。
