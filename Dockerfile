FROM golang:1.15-alpine AS build

WORKDIR /go/src/aci-chatbot
ADD . /go/src/aci-chatbot/
RUN CGO_ENABLED=0 go build -o /bin/aci-chatbot

FROM scratch

COPY --from=build /bin/aci-chatbot /bin/aci-chatbot
EXPOSE 7001

ENTRYPOINT ["/bin/aci-chatbot"]
