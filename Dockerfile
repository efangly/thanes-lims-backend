# syntax=docker/dockerfile:1

# ---- Build stage ----
# glibc (not Alpine/musl) so the cgo binary built here is ABI-compatible with
# the glibc runtime image below - Oracle Instant Client is glibc-only, and a
# musl-built cgo binary won't run against it.
FROM oraclelinux:9 AS build

ARG GO_VERSION=1.26.5
ARG TARGETARCH

RUN dnf install -y gcc git tar gzip xz ca-certificates \
    && dnf clean all

RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${TARGETARCH}.tar.gz" | tar -C /usr/local -xz
ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# godror (Oracle driver) requires cgo. Oracle Instant Client itself is
# dlopen'd by ODPI-C at connect time, not link time, so it isn't needed here
# - only in the runtime stage.
RUN CGO_ENABLED=1 GOOS=linux go build \
    -trimpath -ldflags="-s -w" \
    -o /out/api ./cmd/api

# ---- Runtime stage ----
FROM oraclelinux:9-slim AS runtime

# Oracle Instant Client Basic Lite - required at runtime for godror/ODPI-C to
# connect to the ADB. Installed straight from Oracle's RPM download (no repo
# config needed - the RPM has no external deps beyond glibc/libaio). ldconfig
# registers its libs system-wide via the RPM's /etc/ld.so.conf.d/ drop-in, so
# no LD_LIBRARY_PATH is needed.
ARG INSTANTCLIENT_VERSION=23.26.0.0.0
ARG INSTANTCLIENT_RPM_DIR=2326000
RUN microdnf install -y ca-certificates tzdata curl libaio \
    && curl -fsSL -o /tmp/instantclient-basiclite.rpm \
        "https://download.oracle.com/otn_software/linux/instantclient/${INSTANTCLIENT_RPM_DIR}/oracle-instantclient-basiclite-${INSTANTCLIENT_VERSION}-1.el9.x86_64.rpm" \
    && rpm -ivh /tmp/instantclient-basiclite.rpm \
    && rm -f /tmp/instantclient-basiclite.rpm \
    && microdnf clean all \
    && ldconfig

RUN groupadd -g 10001 app && useradd -u 10001 -g app -M -s /sbin/nologin app

WORKDIR /app

COPY --from=build /out/api ./api
COPY migrations ./migrations

# ADB wallet mount point (populated via k8s Secret volume "adb-wallet" in prod,
# or a bind/named volume for local docker run).
RUN mkdir -p /app/wallet && chown -R app:app /app/wallet
VOLUME ["/app/wallet"]

USER 10001:10001

EXPOSE 8080

ENTRYPOINT ["./api"]
