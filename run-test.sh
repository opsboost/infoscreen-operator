#!/usr/bin/env bash

[[ "$TRACE" ]] && set -x
set -eo pipefail

INSTALL_MINIKUBE="1"
ARCH=amd64

pod_running() {
    if [ "$(kubectl get pods -A | grep "${1}" | awk '{print $4}')" != "Running" ]; then
        echo "1"
    else
        echo "0"
    fi
}

pod_name() {
    local name
    name=$(kubectl get pods -A | grep "${1}" | awk '{print $2}')
    echo "$name"
}

container_name() {
    local arr
    local name
    arr=(${1//:/ })
    name=${arr[0]}
    arr=(${name//\// })
    name=${arr[2]}
    echo "$name"
}

minikube_install() {
    sudo apt install -y podman curl
    curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-${ARCH}
    curl -LO "https://dl.k8s.io/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/${ARCH}/kubectl.sha256"
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/${ARCH}/kubectl"
    echo "$(cat kubectl.sha256)  kubectl" | sha256sum --check
    sudo install minikube-linux-${ARCH} /usr/local/bin/minikube
    sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
    kubectl version --client
}

minikube_deinstall() {
    minikube stop
    minikube delete --all
}

check_dependencies() {
    for i in "${DEPS[@]}"
    do
        if [[ -z $(which "${i}") ]]; then
            error "Could not find ${i}"
            exit 1
        fi
    done
}

main() {
    local image="${1}"

    # Get container name
    local container
    container=$(container_name "${image}")

    # Check for minikube
    # Install if absent and installation enabled
    if [[ -z $(which "minikube") ]]; then
        printf "Could not find minikube"
        if [ "$INSTALL_MINIKUBE" -eq "1" ]; then
            minikube_install
        fi
    fi

    # Check for a running minikube
    if minikube status | grep 'Running'; then
        printf "minikube is already running.\n"
        printf "To continue, stop minikube and restart script\n"
        exit 1
    fi

    # Run minikube
    minikube start --driver=podman --container-runtime=cri-o

    # Deploy
    kubectl create deployment "${container}" --image="${image}"

    # Wait for container to come up
    sleep 30
    kubectl get pods -A -o wide

    # Get pod name
    local pod
    pod=$(pod_name "${container}")

    # Show logs
    kubectl logs "${pod}"

    local r
    r=$(pod_running "${pod}")

    if [ "$r" -eq 0 ]; then
        printf "Deployment successful\n"
    else
        printf "Deployment failed\n"
        local fail=1
    fi

    # Cleanup
    minikube_deinstall

    if [ "$fail" -eq "1" ]; then
        exit 1
    fi
}

main "$@"
