FROM node:latest AS builder

WORKDIR /app/frontend

COPY frontend/package.json frontend/package-lock.json /app/frontend
RUN npm install

COPY frontend /app/frontend
COPY api/v1/spec.yaml /app/api/v1/spec.yaml

RUN npm run build

CMD ["npm", "run", "preview", "--", "--host"]