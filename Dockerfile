#
# Copyright: Pixel Networks <support@pixel-networks.com> 
# Author: Oleg Borodin <oleg.borodin@pixel-networks.com>
#
#
# Build stage
#
FROM golang:1.16.5-alpine AS builder

ENV TARGET=pmdrv
ENV SRCDIR=/go/src/app

WORKDIR $SRCDIR
COPY . .
ENV CGO_ENABLED=0 
RUN go build -o /$TARGET $TARGET.go

#
# Final stage
#
FROM alpine

ENV TARGET=pmdrv
ENV RUNDIR=/
ENV DATADIR=/pmdata

WORKDIR /

COPY --from=pixelcore.azurecr.io/pixctl:latest /pixctl $RUNDIR
COPY --from=builder /$TARGET $RUNDIR/
COPY ./start $RUNDIR/

COPY ./$DATADIR $DATADIR
ADD ./$TARGET.yml $RUNDIR/
ENTRYPOINT ["sh", "/start"]
#EOF

