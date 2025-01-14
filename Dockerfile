FROM node:23-alpine AS frontend
ARG VITE_URL="https://merch.ntuscds.com"
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable
COPY --exclude=backend/ . /app
WORKDIR /app
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile
RUN pnpm run build

FROM golang:1.23-alpine AS backend
RUN apk add --no-cache git make build-base
WORKDIR /app
COPY backend/go.mod backend/go.sum .
RUN go mod download
ADD ./backend /app
ENV CGO_ENABLED=1
RUN go build -o ./ccds-shop .

FROM alpine:3.21
WORKDIR /app
COPY --from=frontend /app/build ./static
COPY --from=backend /app/ccds-shop ./ccds-shop
COPY ./backend/prod.sh .

EXPOSE 8080
CMD ["/app/prod.sh"]
