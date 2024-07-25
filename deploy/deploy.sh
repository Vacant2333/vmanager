kind delete cluster --name kind5
kind create cluster --name kind5 --config deploy/kind-config.yaml


kubectl label nodes kind5-control-plane node.kubernetes.io/capacity=on-demand
kubectl label nodes kind5-worker node.kubernetes.io/capacity=on-demand
kubectl label nodes kind5-worker2 node.kubernetes.io/capacity=on-demand

kubectl label nodes kind5-worker3 node.kubernetes.io/capacity=spot
kubectl label nodes kind5-worker4 node.kubernetes.io/capacity=spot

make images

kind load docker-image --name kind5 vacantsh/webhook-manager:1.0

kubectl apply -f deploy/namespace.yaml
kubectl apply -f deploy/webhook-manager.yaml
kubectl apply -f deploy/webhooks.yaml

