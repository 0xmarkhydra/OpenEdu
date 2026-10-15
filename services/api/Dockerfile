FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /openedu-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /openedu-bootstrap ./cmd/bootstrap

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=build /openedu-api /openedu-api
COPY --from=build /openedu-bootstrap /openedu-bootstrap
ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/openedu-api"]
