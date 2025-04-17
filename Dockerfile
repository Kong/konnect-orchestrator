FROM gcr.io/distroless/base-debian11:nonroot

WORKDIR /app

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["./koctl"]

COPY koctl /app