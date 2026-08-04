FROM ubuntu:24.04 AS builder

ARG DEBIAN_FRONTEND=noninteractive
ENV TZ=Etc/UTC

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        build-essential \
        gcc-13 \
        g++-13 \
        cmake \
        git \
        curl \
        ca-certificates \
        pkg-config \
        ninja-build \
        python3 \
        zip \
        unzip \
        tar \
        bison \
        flex \
        autoconf \
        automake \
        libtool \
    && rm -rf /var/lib/apt/lists/*

ENV CC=/usr/bin/gcc-13
ENV CXX=/usr/bin/g++-13

SHELL ["/bin/bash", "-c"]

# vcpkg.json 의 builtin-baseline 과 같은 커밋을 얕게 가져온다. tip 을 clone 하면 baseline 커밋이 없다.
ARG VCPKG_BASELINE=40f3c709db80acf154ac4b17a1f83c564ebd022e
WORKDIR /opt/vcpkg
RUN git init -q . \
    && git remote add origin https://github.com/microsoft/vcpkg.git \
    && git fetch -q --depth 1 origin "$VCPKG_BASELINE" \
    && git checkout -q FETCH_HEAD \
    && ./bootstrap-vcpkg.sh -disableMetrics
ENV VCPKG_ROOT=/opt/vcpkg
ENV VCPKG_INSTALLED_DIR=/opt/vcpkg_installed

# resolve triplet based on container architecture unless explicitly provided
ARG VCPKG_TRIPLET=
RUN triplet="$VCPKG_TRIPLET"; \
    if [ -z "$triplet" ]; then \
      arch=$(dpkg --print-architecture); \
      case "$arch" in \
        amd64) triplet="x64-linux" ;; \
        arm64) triplet="arm64-linux" ;; \
        *) triplet="${arch}-linux" ;; \
      esac; \
    fi; \
    echo "$triplet" > /opt/vcpkg-triplet

WORKDIR /workspace

# 의존성 설치를 별도 레이어로 둔다. 소스만 바뀌면 포트 전체를 다시 빌드하지 않는다.
COPY vcpkg.json ./
RUN "${VCPKG_ROOT}/vcpkg" install \
        --triplet "$(cat /opt/vcpkg-triplet)" \
        --x-install-root="${VCPKG_INSTALLED_DIR}"

COPY . .

RUN rm -rf build \
    && cmake -G Ninja -S . -B build \
        -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_TOOLCHAIN_FILE=${VCPKG_ROOT}/scripts/buildsystems/vcpkg.cmake \
        -DVCPKG_TARGET_TRIPLET="$(cat /opt/vcpkg-triplet)" \
        -DVCPKG_INSTALLED_DIR="${VCPKG_INSTALLED_DIR}" \
        -DVCPKG_FEATURE_FLAGS=manifests \
        -DCMAKE_CXX_FLAGS="-include cstring" \
        -DBUILD_SHARED_LIBS=OFF \
    && cmake --build build --config Release --parallel

FROM ubuntu:24.04 AS runtime

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

RUN userdel -r ubuntu >/dev/null 2>&1 || true; \
    groupadd -g 1000 appuser \
    && useradd -u 1000 -g 1000 -m -s /usr/sbin/nologin appuser \
    && chown -R appuser:appuser /opt/app

USER appuser

EXPOSE 9876
CMD ["/opt/app/bin/MainServer"]
