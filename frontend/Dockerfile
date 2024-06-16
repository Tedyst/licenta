FROM node:latest AS builder

WORKDIR /app/frontend

COPY frontend/package.json frontend/package-lock.json /app/frontend
RUN npm install

COPY frontend /app/frontend
COPY api/v1/spec.yaml /app/api/v1/spec.yaml

ENV HOST=0.0.0.0
ENV PORT=3000
ENV PROTOCOL_HEADER=x-forwarded-proto
ENV HOST_HEADER=x-forwarded-host

RUN npm run build

CMD ["node", "build"]