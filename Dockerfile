FROM alpine

LABEL authors="lwnmengjing"

COPY ./pushMessage /app/pushMessage

ENTRYPOINT ["/app/admipushMessage"]