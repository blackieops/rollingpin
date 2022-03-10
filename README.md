# 🧑‍🍳 `rollingpin`

`rollingpin` is a small Go infrastructure service that receives webhooks from a
container regsitry and updates corresponding Kubernetes deployments to
facilitate minimal continuous delivery without complex CI/CD pipelines.

## Usage

`rollingpin` is first and foremost a Kubernetes service, so the only supported
usage is as a container in a Kubernetes pod.

### Authorization

As `rollingpin` modifies Kubernetes resources, it needs to be authorized to
take certain actions against the Kubernetes API. You'll need to set up a
Service Account and a couple RBAC policies to allow this.

First, a service account:

```yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rollingpin
```

Then, for each namespace you want `rollingpin` to access, you'll need to create
a `Role` and `RoleBinding`. It is highly recommended to be as strict as
possible, and to avoid using a `ClusterRole` if at all possible.

The only permissions needed are to `get` and `update` deployments.

```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: default
  name: rollingpin-deploy
rules:
- apiGroups: ["apps/v1"]
  resources: ["deployments"]
  verbs: ["get", "update"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: rollingpin
  namespace: default
subjects:
- kind: ServiceAccount
  name: rollingpin
  apiGroup: ""
roleRef:
  kind: Role
  name: rollingpin-deploy
  apiGroup: rbac.authorization.k8s.io
```

### Deployment

`rollingpin` is a stateless service; deploy it as you would any other container
on Kubernetes. Just ensure you set the `serviceAccountName` in your pod spec to
match the Service Account you created for `rollingpin`.

Containers can either by built yourself, or pulled from our registry:

```
cr.b8s.dev/library/rollingpin:v1.0.0
```
