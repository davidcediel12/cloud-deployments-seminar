# Two stages: the first has the Go compiler and builds the binary,
# the second contains only that binary.

# ---------- Stage 1: build ----------
FROM golang:1.26 AS build
WORKDIR /src

# Dependencies first: this layer is cached until go.mod/go.sum change.
COPY go.mod go.sum ./
RUN go mod download

# Then the code, which changes often.
COPY . .

# Static binary: no system libraries needed at runtime.
RUN CGO_ENABLED=0 go build -o /bin/api .

# ---------- Stage 2: run ----------
# No compiler, no shell, runs as a non-root user.
FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /bin/api /api
EXPOSE 8080
ENTRYPOINT ["/api"]
