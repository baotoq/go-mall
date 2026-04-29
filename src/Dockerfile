ARG SERVICE

FROM golang:1.26-alpine AS build
ARG SERVICE
WORKDIR /workspace
RUN apk add --no-cache gcc musl-dev
COPY . ./src/
RUN cd src/${SERVICE} && CGO_ENABLED=1 go build -ldflags="-w -s" -o /out/app .

FROM alpine:3.20
ARG SERVICE
ENV SERVICE=${SERVICE}
RUN apk add --no-cache ca-certificates tzdata libgcc
WORKDIR /app
COPY --from=build /out/app ./app
COPY --from=build /workspace/src/${SERVICE}/etc ./etc
ENTRYPOINT ["/bin/sh", "-c", "/app/app -f etc/${SERVICE}-api.yaml"]
