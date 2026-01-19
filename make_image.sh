export VERSION='0.36.2.3'
export GOFLAGS="-ldflags=-X=sigs.k8s.io/karpenter/pkg/operator.Version=${VERSION}"
export SOURCE_DATE_EPOCH="$(git log -1 --format='%ct')"
export KO_DATA_DATE_EPOCH="$(git log -1 --format='%ct')"
export KO_DOCKER_REPO=avant.jfrog.io/shared-images/karpenter
ko publish -B --platform=linux/amd64 -t "${VERSION}" ./cmd/controller



# ko publish -B --platform=linux/amd64,linux/arm64 -t "${VERSION}" ./cmd/controller