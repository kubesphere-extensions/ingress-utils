# Copyright 2024 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

CRD_OPTIONS ?= "crd:allowDangerousTypes=true"
MANIFESTS="gateway/v1alpha2"

REPO ?= ${REPO:-kubesphere}
TAG ?= ${TAG:-latest}
IMG := ${REPO}/ingress-utils-controller-manager:${TAG}

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

manifests: ;$(info $(M)...Begin to generate manifests e.g. CRD etc..)
	hack/generate_manifests.sh ${CRD_OPTIONS} ${MANIFESTS}

deepcopy: ;$(info $(M)...Begin to deepcopy.)
	hack/generate_manifests.sh ${CRD_OPTIONS} ${MANIFESTS} "deepcopy"

.PHONY: docker-build
docker-build: ;$(info $(M)...Begin to build controller image.)
	docker build -t ${IMG} -f ./build/controller-manager/Dockerfile .

.PHONY: docker-push
docker-push: ;$(info $(M)...Begin to push controller image.)
	docker push ${IMG}

container-push: | docker-build docker-push; $(info $(M)...Begin to build and push controller image.)

container-cross-push: ; $(info $(M)...Begin to build and push container images for multiple platforms.)
	docker buildx build -t ${IMG} -f ./build/controller-manager/Dockerfile --push --platform linux/amd64,linux/arm64 .
	