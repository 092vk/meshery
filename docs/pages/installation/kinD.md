---
layout: page
title: KinD
permalink: installation/kind
---

# Quick Start with KinD

## Installation and usage

On Mac / Linux via Homebrew:
```
brew install kind
```

On macOS / Linux:
```
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.8.1/kind-$(uname)-amd64
chmod +x ./kind
mv ./kind /some-dir-in-your-PATH/kind
```

Create cluster using KinD
```
kind create cluster --name kind --wait 300s
Creating cluster "kind" ...
 • Ensuring node image (kindest/node:v1.17.0) 🖼  ...
 ✓ Ensuring node image (kindest/node:v1.17.0) 🖼
 • Preparing nodes 📦   ...
 ✓ Preparing nodes 📦 
 • Writing configuration 📜  ...
 ✓ Writing configuration 📜
 • Starting control-plane 🕹️  ...
 ✓ Starting control-plane 🕹️
 • Installing CNI 🔌  ...
 ✓ Installing CNI 🔌
 • Installing StorageClass 💾  ...
 ✓ Installing StorageClass 💾
 • Waiting ≤ 5m0s for control-plane = Ready ⏳  ...
 ✓ Waiting ≤ 5m0s for control-plane = Ready ⏳
 • Ready after 59s 💚
Set kubectl context to "kind-kind"
You can now use your cluster with:

kubectl cluster-info --context kind-kind

Not sure what to do next? 😅 Check out https://kind.sigs.k8s.io/docs/user/quick-start/
```

By default, the cluster access configuration is stored in ${HOME}/.kube/config if $KUBECONFIG environment variable is not set. You can set the `KIUBECONFIG` environment command below:
```
export KUBECONFIG=${HOME}/.kube/config
```
Use the command below check the connection of the cluster and make sure the cluster you connected what's the cluster was created by KinD:
```
kubectl cluster-info --context kind-kind
```

To delete your cluster use: 
```
kind delete cluster --name kind
```

<a name="helm"></a>
## Using Helm

### Helm v3

We strongly recommend to use Helm v3, because of the version not included the Tiller(https://helm.sh/blog/helm-3-preview-pt2/#helm) component anymore. It’s more light and safe.

Run the following:
```
$ git clone https://github.com/layer5io/meshery.git; cd meshery
$ kubectl create namespace meshery
$ helm install meshery --namespace meshery install/kubernetes/helm/meshery
```

* **NodePort** - If your cluster does not have an Ingress Controller or a load balancer, then use NodePort to expose Meshery and that can be modify under the chart `values.yaml`:

```
service:
  type: NodePort
  port: 8080
  annotations: {}
```
