FROM node:20-alpine AS builder

WORKDIR /app
COPY package.json ./
RUN npm install
COPY tsconfig.json ./
COPY src ./src
RUN npm run build

# ── runtime ──────────────────────────────────────────────
FROM node:20-alpine

WORKDIR /app

# production 의존성만 설치
COPY package.json ./
RUN npm install --omit=dev

COPY --from=builder /app/dist ./dist

# 크리덴셜 파일 저장 경로 (PVC 마운트 포인트)
VOLUME ["/data"]

# non-root 실행
RUN addgroup -S botgroup && adduser -S botuser -G botgroup
USER botuser

ENV NODE_ENV=production \
    CREDENTIAL_PATH=/data/credential.json

CMD ["node", "dist/index.js"]
