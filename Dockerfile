FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./backend/cmd/api
RUN CGO_ENABLED=0 go build -o /out/migrate ./backend/cmd/migrate

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /bin/api
COPY --from=build /out/migrate /bin/migrate
COPY backend/migrations /migrations
COPY openapi/openapi.yaml /openapi/openapi.yaml
EXPOSE 8080
ENTRYPOINT ["/bin/api"]
