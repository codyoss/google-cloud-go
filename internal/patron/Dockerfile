# Copyright 2025 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
FROM marketplace.gcr.io/google/debian12:latest

# Set environment variables for the Go version and installation
ENV GO_VERSION 1.23.0
ENV PROTOC_VERSION 25.7
ENV GO_PROTOC_PLUGIN_VERSION 1.36.6
ENV GAPIC_GENERATOR_VERSION 0.53.1

# Set the PATH. Note: /root/go/bin is for tools installed by the root user.
# The user's GOPATH will be different.
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

# Install linux tooling
RUN apt-get update && \
    apt-get install -y wget ca-certificates git unzip openssh-client \
    build-essential curl && \
    rm -rf /var/lib/apt/lists/*

# Install Go
RUN wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz -O go.tar.gz && \
    tar -C /usr/local -xzf go.tar.gz && \
    rm go.tar.gz && \
    go version

# Install protoc
RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip && \
    unzip protoc-${PROTOC_VERSION}-linux-x86_64.zip && \
    which protoc && protoc --version

# Install Go tools.
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v${GO_PROTOC_PLUGIN_VERSION} && \
    go install golang.org/x/lint/golint@latest && \
    go install golang.org/x/tools/cmd/goimports@latest && \
    go install honnef.co/go/tools/cmd/staticcheck@latest && \
    go install github.com/googleapis/gapic-generator-go/cmd/protoc-gen-go_gapic@v${GAPIC_GENERATOR_VERSION}

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -o patron

ENTRYPOINT ["/app/patron"]