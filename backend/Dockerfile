# ---- build stage: compile a static Go binary ----
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/tasker-api .

# ---- runtime stage: tiny image, no shell, non-root ----
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/tasker-api /tasker-api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/tasker-api"]
