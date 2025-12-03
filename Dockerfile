# Multi-stage build for RealTimeMessageChat
FROM ubuntu:22.04 AS builder

# Prevent interactive prompts
ENV DEBIAN_FRONTEND=noninteractive

# Install dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    gcc \
    g++ \
    make \
    cmake \
    git \
    curl \
    wget \
    zip \
    unzip \
    tar \
    pkg-config \
    autoconf \
    automake \
    libtool \
    bison \
    flex \
    linux-libc-dev \
    libssl-dev \
    python3 \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

# Install vcpkg
WORKDIR /opt
RUN git clone https://github.com/Microsoft/vcpkg.git
WORKDIR /opt/vcpkg
RUN ./bootstrap-vcpkg.sh

# Set vcpkg environment
ENV VCPKG_ROOT=/opt/vcpkg
ENV PATH="${VCPKG_ROOT}:${PATH}"

# Copy source code
WORKDIR /app
COPY . .

# Initialize git submodules
RUN git submodule update --init --recursive || true

# Build the project
RUN mkdir -p build && cd build && \
    cmake .. \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_TOOLCHAIN_FILE=/opt/vcpkg/scripts/buildsystems/vcpkg.cmake \
    -DCMAKE_C_COMPILER=/usr/bin/gcc \
    -DCMAKE_CXX_COMPILER=/usr/bin/g++ && \
    cmake --build . --target MainServer MainServerConsumer -j$(nproc)

# Runtime stage for MainServer
FROM ubuntu:22.04 AS mainserver

RUN apt-get update && apt-get install -y \
    libssl3 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy MainServer binary and configuration
COPY --from=builder /app/build/out/MainServer /app/
COPY --from=builder /app/build/out/main_server_configurations.json /app/

# Copy entrypoint script
COPY docker/mainserver-entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

EXPOSE 9876

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["/app/MainServer"]

# Runtime stage for MainServerConsumer
FROM ubuntu:22.04 AS consumer

RUN apt-get update && apt-get install -y \
    libssl3 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy MainServerConsumer binary and configuration
COPY --from=builder /app/build/out/MainServerConsumer /app/
COPY --from=builder /app/build/out/main_server_consumer_configurations.json /app/

# Copy entrypoint script
COPY docker/consumer-entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["/app/MainServerConsumer"]
