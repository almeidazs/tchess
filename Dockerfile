# Image published by GoReleaser (see .goreleaser.yaml): it copies the
# binary already built for each platform, so this file is not meant for a
# plain `docker build`. Run it with a terminal attached:
#
#   docker run --rm -it ghcr.io/almeidazs/tchess
#
# Preferences and games live in /root/.tchess; mount a volume there to keep
# them: -v tchess:/root/.tchess
FROM scratch
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/tchess /usr/bin/tchess
ENV HOME=/root TERM=xterm-256color COLORTERM=truecolor
ENTRYPOINT ["/usr/bin/tchess"]
