FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/fit-metrics ./cmd/fit-metrics
RUN CGO_ENABLED=0 go build -o /out/fit-structure ./cmd/fit-structure
RUN CGO_ENABLED=0 go build -o /out/fit-features ./cmd/fit-features
RUN CGO_ENABLED=0 go build -o /out/fit-summary ./cmd/fit-summary
RUN CGO_ENABLED=0 go build -o /out/fit-chat ./cmd/fit-chat

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/fit-metrics /fit-metrics
COPY --from=build /out/fit-structure /fit-structure
COPY --from=build /out/fit-features /fit-features
COPY --from=build /out/fit-summary /fit-summary
COPY --from=build /out/fit-chat /fit-chat
USER nonroot:nonroot
CMD ["/fit-metrics"]
