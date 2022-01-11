FROM golang:1.15-alpine AS build

WORKDIR /go/src/chatbot
ADD . /go/src/chatbot/
RUN CGO_ENABLED=0 go build -o /bin/chatbot

FROM scratch

COPY --from=build /bin/chatbot /bin/chatbot
EXPOSE 7001

ENTRYPOINT ["/bin/chatbot"]
