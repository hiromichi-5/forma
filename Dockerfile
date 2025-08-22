FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN go build -o /out/api ./cmd/api

FROM gcr.io/distroless/base
COPY --from=build /out/api /bin/api
EXPOSE 8080
ENTRYPOINT ["/bin/api"]
