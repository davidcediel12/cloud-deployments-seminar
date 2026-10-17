# Hands-on 1: write the Dockerfile for the API.
# Replace each TODO with the instruction in its hint. Read the "why": that's the point of the exercise.
# Stuck? `git stash` + `git switch step-1-containerized` has the solution.
#
# Two stages: the first has the Go compiler and builds the binary,
# the second contains only that binary.

# ---------- Stage 1: build ----------

# TODO 1: start from the official Go image and name this stage "build".
#   Hint: FROM golang:1.26 AS build
#   Why:  "AS build" lets the second stage copy files from this one.

# TODO 2: set the working directory inside the image.
#   Hint: WORKDIR /src
#   Why:  following instructions run in (and copy into) this folder.

# TODO 3: copy only the dependency file, then download dependencies.
#   Hint: COPY go.mod ./
#         RUN go mod download
#   Why:  layer cache. Dependencies change rarely, so this layer is reused
#         and code changes don't trigger a new download.

# TODO 4: copy the rest of the source code.
#   Hint: COPY . .
#   Why:  code changes often, so it goes after the dependencies.

# TODO 5: build a static Linux binary.
#   Hint: RUN CGO_ENABLED=0 go build -o /bin/api .
#   Why:  CGO_ENABLED=0 makes a binary with no system libraries,
#         so it runs on an almost empty base image.

# ---------- Stage 2: run ----------

# TODO 6: start a new, minimal stage that runs as a non-root user.
#   Hint: FROM gcr.io/distroless/static-debian13:nonroot
#   Why:  no compiler, no shell, no root: smaller image, smaller attack surface.

# TODO 7: copy the binary from the "build" stage.
#   Hint: COPY --from=build /bin/api /api
#   Why:  only the result of stage 1 ends up in the final image.

# TODO 8: document the port the app listens on.
#   Hint: EXPOSE 8080
#   Why:  documentation only; `docker run -p` still publishes the port.

# TODO 9: set the command that starts the API.
#   Hint: ENTRYPOINT ["/api"]
#   Why:  JSON form runs the binary directly (there is no shell in this image).
