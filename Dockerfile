FROM scratch
COPY docker-root/ /
WORKDIR /root
ENTRYPOINT ["/bin/dfpp"]