# Run the resulting image with the following:
#
# docker run -it \
#   -v /tmp:/tmp \
#   -v $HOME/.aws:/home/s-vidzou/.aws \
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
RUN adduser --disabled-password --uid 1000 s-vidzou
WORKDIR /vidzou
COPY --from=build-env /src/main /vidzou/
COPY --from=build-env /src/templates /vidzou/templates
RUN chown -R s-vidzou /vidzou
USER s-vidzou
ENTRYPOINT ./main
