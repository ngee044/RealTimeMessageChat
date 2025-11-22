FROM ubuntu:22.04 AS builder

ARG DEBIAN_FRONTEND=noninteractive
ENV TZ=Etc/UTC

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        build-essential \
        cmake \
        git \
        curl \
        ca-certificates \
        pkg-config \
        ninja-build \
        zip \
        unzip \
        tar \
        bison \
        flex \
        autoconf \
        automake \
        libtool \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /opt
RUN git clone --depth 1 https://github.com/microsoft/vcpkg.git
WORKDIR /opt/vcpkg
RUN ./bootstrap-vcpkg.sh -disableMetrics
ENV VCPKG_ROOT=/opt/vcpkg

WORKDIR /workspace
COPY . .

# resolve triplet based on container architecture unless explicitly provided
ARG VCPKG_TRIPLET=
SHELL ["/bin/bash", "-c"]

RUN triplet="$VCPKG_TRIPLET"; \
    if [ -z "$triplet" ]; then \
      arch=$(dpkg --print-architecture); \
      case "$arch" in \
        amd64) triplet="x64-linux" ;; \
        arm64) triplet="arm64-linux" ;; \
        *) triplet="${arch}-linux" ;; \
      esac; \
    fi; \
    rm -rf build \
    && cmake -G Ninja -S . -B build \
        -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_TOOLCHAIN_FILE=${VCPKG_ROOT}/scripts/buildsystems/vcpkg.cmake \
        -DVCPKG_TARGET_TRIPLET="$triplet" \
        -DVCPKG_FEATURE_FLAGS=manifests \
        -DBUILD_SHARED_LIBS=OFF \
    && cmake --build build --config Release --parallel

FROM ubuntu:22.04 AS runtime

ARG DEBIAN_FRONTEND=noninteractive
ENV TZ=Etc/UTC

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        libstdc++6 \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /opt/app/logs

WORKDIR /opt/app/bin
COPY --from=builder /workspace/build/out/ ./
COPY docker/config/main_server_configurations.json ./main_server_configurations.json
COPY docker/config/main_server_consumer_configurations.json ./main_server_consumer_configurations.json

EXPOSE 9876
CMD ["/opt/app/bin/MainServer"]
