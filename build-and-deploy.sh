TAG=$1

docker build -t "controller-reconciler:$TAG" . 

kind load docker-image "controller-reconciler:$TAG"

make deploy IMG="controller-reconciler:$TAG"
