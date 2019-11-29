# We use the ubuntu 18.04 base image, as `apt` makes it easy to install
# `youtube-dl` dependencies (i.e. ffmpeg and python3).
FROM ubuntu:18.04

RUN apt update
RUN apt install -y ffmpeg python3 wget

# `youtube-dl` will look for `/usr/bin/env python`. Without this symlink,
# `/usr/bin/env python` will fail. We could also address this problem via
# installing python2 or potentially using `update-alternatives`. We're trying to
# avoid the first as the py2 EOL is quickly approaching and the
# `update-alternatives` approach seemed more trouble than its worth given we're
# building a container and will only ever install one python version.
RUN ln -s /usr/bin/python3 /usr/bin/python

# We decided NOT to create a specific `youtube-dl` user to run the ./youtube-dl
# process. Instead, we expect whoever is running a container from this image to
# specify a proper user id. It's easier if whoever runs the container picks the
# user id, as, with that level of control, they can ensure any files created by
# youtube-dl have the proper user on the host.
#
# Regardless of the specific user running the container, our configuration in
# `youtube-dl.conf` ensures all  all downloaded articats will be stored in `/downloads`.
# Since the youtube-dl binary could be running as any user, `/downloads` should
# be world writable.
#
# When running the container, we expect the entity running the container to
# mount a directory (ex. /tmp/downloads) into `/downloads`, so the host system
# can access the downloaded files.
RUN mkdir /downloads
RUN chmod 777 /downloads

ARG youtube_dl_version=2019.11.28

# Download and validate the youtube-dl binary.
RUN wget https://github.com/ytdl-org/youtube-dl/releases/download/$youtube_dl_version/youtube-dl
RUN wget https://github.com/ytdl-org/youtube-dl/releases/download/$youtube_dl_version/SHA2-512SUMS
RUN sha512sum --ignore-missing -c SHA2-512SUMS

# As our container could be running as any user, we want any user to be able to
# execute the youtube-dl binary.
RUN chmod 747 youtube-dl

COPY youtube-dl.conf /etc/youtube-dl.conf

ENTRYPOINT ["./youtube-dl"]
CMD ["--help"]
