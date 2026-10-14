FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /freeeducation-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /freeeducation-bootstrap ./cmd/bootstrap

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=build /freeeducation-api /freeeducation-api
COPY --from=build /freeeducation-bootstrap /freeeducation-bootstrap
ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/freeeducation-api"]
