# While this Dockerfile exists, I'm not actually sure if initially I want to run
# the app in a Dockerfile, or instead just build a go binary and deploy it via a
# simple systemd service... tbd.
#
# Run the resulting image with the following:
#
# Add to Makefile.
#
# docker run -it \
#   -v /tmp:/tmp \
#   -v $HOME/.aws:/home/s-go/.aws \
#   -v /var/run/docker.sock:/var/run/docker.sock \
#   --group-add 999 \ # The docker group...
#   -p 8080:8080 \
#   IMAGE

# I think alpine makes sense as a default image...
FROM golang:1.13-alpine AS build-env
# TODO: Add appropriate metadata, etc...
RUN apk --no-cache add build-base git bzr mercurial gcc
ADD . /src/
RUN cd /src && go build -o main

FROM alpine
# The `uid` must be the same as the user which owns the ~/.aws directory...
# Or the ~/.aws directory must be group readable.
RUN adduser --disabled-password --uid 1000 s-go
WORKDIR /app
COPY --from=build-env /src/main /app/
COPY --from=build-env /src/templates /app/templates
RUN chown -R s-go /app
USER s-go
ENTRYPOINT ./main
