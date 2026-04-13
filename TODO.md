# TODO — 카카오 봇 서버 사이드 전환

메신저봇R(Android) 의존성을 제거하고, node-kakao 기반 서버 사이드 봇으로 전환한다.

## 배경

현재 구조는 Android 기기의 카카오 알림을 메신저봇R이 가로채는 방식이라,
방 트래픽이 높을 때 알림 스로틀링 + JS 단일 스레드 블로킹으로 봇이 무반응 상태가 된다.

목표 구조:
```
[카카오 서버]
     │ Loco WebSocket (상시 연결)
     ▼
[kakao-client pod]  ← 신규
  Node.js + node-kakao
     │ HTTP (cluster 내부)
     ▼
[yangobot pod]      ← 변경 없음
  http://yangobot/api/v1/...
```

---

## Phase 1 — 로컬 구현

- [x] `kakao-client/` 디렉토리 생성, Node.js 프로젝트 초기화
- [x] node-kakao 4.5.0 의존성 추가, TypeScript 설정
- [x] 메시지 수신 핸들러 구현 (`src/bot.ts`)
  - messagerbot.js의 정규식 패턴 포팅 (`.군장`, `.스펙`, `.캐릭터`, `.원정대`, `.ㄱㅁ4/8`)
  - yangobot API 호출 모듈 분리 (`src/api.ts`)
- [x] 인증 플로우 구현 (`src/index.ts`)
  - AuthApiClient → OAuthCredential 획득
  - 최초 기기 등록 (DEVICE_NOT_REGISTERED → 패스코드 → registerDevice)
  - 크리덴셜 파일 영속화 (`/data/credential.json`)
- [x] 지수 백오프 자동 재연결 로직 구현
- [x] 실행 중 방 추가 명령어 구현 — 세션 끊김 없이 오픈채팅 참여
  - `.방추가 <url>` — 비밀번호 없는 방
  - `.방추가 <url> <비밀번호>` — 입장 비밀번호 있는 방
  - 관리자(ADMIN_USER_IDS) 전용, 이미 참여 중이면 중복 방지
- [x] **카카오 봇 전용 계정**: woohalabs@gmail.com
- [ ] **관리자 userId 확인 및 ADMIN_USER_IDS 설정** ← 직접 해야 함
  - 봇 실행 후 ADMIN_USER_IDS 미설정 상태에서 아무 `.` 명령어 입력
  - 로그에서 `senderId=...` 확인 후 `.env`에 입력
- [ ] 로컬 end-to-end 테스트
  1. `cd kakao-client && cp .env.example .env` → 계정/관리자 정보 입력
  2. `npm install && npm run dev`
  3. `DEVICE_NOT_REGISTERED` 에러 시 이메일로 온 패스코드를 `KAKAO_PASSCODE=숫자` 설정 후 재실행
  4. 실제 오픈채팅방에서 `.군장 아비투스` 등 테스트
  5. `.방추가 <url>` 및 `.방추가 <url> <비밀번호>` 동작 확인

---

## Phase 2 — 컨테이너화

- [x] `kakao-client/Dockerfile` 작성 (multi-stage, non-root)
- [x] 크리덴셜 영속화 방식 결정: PVC 마운트 (`/data/credential.json`)
- [ ] Docker 이미지 빌드 및 로컬 컨테이너 동작 확인 ← 로컬 테스트 후 진행
- [ ] DockerHub 리포지토리 생성 (`ggorockee/yangobot-kakao-client`)

---

## Phase 3 — k8s 배포

- [x] `k8s/kakao-client-deployment.yaml` 작성
  - `replicas: 1` + `strategy: Recreate` (세션 충돌 방지)
  - `YANGOBOT_API_URL=http://yangobot` env 주입
  - `ADMIN_USER_IDS`, `KAKAO_EMAIL`, `KAKAO_PASSWORD` → Secret 마운트
  - `/data` → PVC 마운트 (크리덴셜 파일 보존)
- [ ] `kakao-client-secrets` Secret 생성 ← 직접 해야 함
  ```bash
  kubectl create secret generic kakao-client-secrets \
    --from-literal=kakao-email=woohalabs@gmail.com \
    --from-literal=kakao-password=<비밀번호> \
    --from-literal=admin-user-ids=<userId>
  ```
- [ ] infra repo `charts/helm/prod/kakao-client/values.yaml` 추가
- [ ] ArgoCD 앱 등록 (또는 기존 yangobot 앱에 포함)
- [ ] 클러스터 배포 후 로그 확인

---

## Phase 4 — CI/CD 연동

- [ ] `.github/workflows/ci.yml`에 `kakao-client` 빌드/푸시 job 추가
  - `kakao-client/` 경로 변경 시에만 트리거 (`paths` 필터)
- [ ] `update_values.py` 스크립트에 `kakao-client` image tag 업데이트 로직 추가
- [ ] infra repo PR → ArgoCD 자동 sync 확인

---

## Phase 5 — 안정화 및 정리

- [ ] 메신저봇R과 병행 운영하며 서버 봇 안정성 확인 (최소 1주)
- [ ] Liveness probe 구성 — WebSocket 연결 상태 기반 헬스체크
- [ ] 안정 확인 후 메신저봇R(Android 기기) 종료
- [ ] node-kakao 버전 고정 및 업스트림 breaking change 모니터링 방법 정리

---

## 방 추가 사용법 (운영 중 세션 끊김 없이)

```
# 비밀번호 없는 방
.방추가 https://open.kakao.com/o/gXXXXXX

# 입장 비밀번호 있는 방
.방추가 https://open.kakao.com/o/gXXXXXX 1234

# 이미 참여 중인 방이면 "이미 참여 중" 응답
```

관리자만 사용 가능. 봇 계정이 방에 직접 참여하므로 폰 로그인 불필요.

---

## 주의사항

- `kakao-client`는 `replicas: 1` 고정. 동일 Kakao 계정 2개 pod → 세션 충돌로 둘 다 끊김.
- node-kakao는 비공식 클라이언트. 카카오 프로토콜 업데이트 시 동작이 깨질 수 있다.
- 봇 계정이 어뷰징으로 판단되면 카카오 측에서 계정 제재 가능.
- 입장 비밀번호는 환경변수나 코드에 저장하지 않음. 명령어로 실시간 전달.
