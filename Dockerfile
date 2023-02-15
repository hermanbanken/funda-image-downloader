FROM golang as build
WORKDIR /workdir
COPY ./ /workdir/
ENV CGO_ENABLED=0
RUN cd /workdir &&  go build -o /server ./

FROM scratch
COPY --from=build /server /server
ENTRYPOINT [ "/server" ]
