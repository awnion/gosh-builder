FROM alpine as builder
FROM debian as builder
FROM builder
RUN ls -la /
