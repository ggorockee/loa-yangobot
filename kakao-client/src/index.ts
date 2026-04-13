/**
 * kakao-client — 카카오 Loco 프로토콜 기반 서버 사이드 봇
 *
 * 실행 흐름:
 *  1. AuthApiClient로 이메일/비번 → OAuthCredential 획득 (최초 1회, 이후 토큰 재사용)
 *  2. TalkClient.login(credential) → Loco WebSocket 연결
 *  3. client.on('chat') → 메시지 수신 → yangobot API 호출 → reply
 *  4. 연결 끊김 시 지수 백오프로 자동 재연결
 *
 * 최초 실행 시 기기 등록 필요:
 *  - DEVICE_NOT_REGISTERED(-100) 에러 발생 → 이메일로 패스코드 발송
 *  - 환경변수 KAKAO_PASSCODE에 패스코드 입력 후 재실행하면 기기 등록 완료
 */

import * as fs from 'fs';
import * as path from 'path';
import * as readline from 'readline';
import { AuthApiClient, KnownAuthStatusCode, TalkClient } from 'node-kakao';
import type { OAuthCredential } from 'node-kakao';
import { handleChat } from './bot';

// ── 환경변수 ───────────────────────────────────────────────
const KAKAO_EMAIL    = requireEnv('KAKAO_EMAIL');
const KAKAO_PASSWORD = requireEnv('KAKAO_PASSWORD');
const DEVICE_NAME    = process.env.DEVICE_NAME    ?? 'yangobot';
const DEVICE_UUID    = process.env.DEVICE_UUID    ?? deriveUUID(KAKAO_EMAIL);
const KAKAO_PASSCODE = process.env.KAKAO_PASSCODE ?? '';

// 기기 등록 후 발급된 크리덴셜을 파일에 저장해 재사용
const CREDENTIAL_PATH = process.env.CREDENTIAL_PATH ?? '/data/credential.json';

// 재연결 설정
const RECONNECT_BASE_MS  = 5_000;
const RECONNECT_MAX_MS   = 5 * 60_000;

// ── 메인 ──────────────────────────────────────────────────
main().catch((err) => {
  console.error('[main] fatal error:', err);
  process.exit(1);
});

async function main(): Promise<void> {
  const credential = await authenticate();
  await connectWithRetry(credential);
}

// ── 인증 ──────────────────────────────────────────────────
async function authenticate(): Promise<OAuthCredential> {
  // 저장된 크리덴셜이 있으면 재사용
  const saved = loadCredential();
  if (saved) {
    console.log('[auth] Using saved credential');
    return saved;
  }

  const authClient = await AuthApiClient.create(DEVICE_NAME, DEVICE_UUID);

  // 기기 미등록 상태 처리
  if (KAKAO_PASSCODE) {
    console.log('[auth] Registering device with passcode...');
    const regResult = await authClient.registerDevice(
      { email: KAKAO_EMAIL, password: KAKAO_PASSWORD },
      KAKAO_PASSCODE,
      true,
    );
    if (!regResult.success && regResult.status !== KnownAuthStatusCode.SUCCESS_SAME_USER) {
      throw new Error(`[auth] Device registration failed: status=${regResult.status}`);
    }
    console.log('[auth] Device registered');
  }

  console.log('[auth] Logging in...');
  const loginResult = await authClient.login(
    { email: KAKAO_EMAIL, password: KAKAO_PASSWORD },
    true, // forced — 다른 기기 세션 강제 종료
  );

  if (!loginResult.success) {
    if (loginResult.status === KnownAuthStatusCode.DEVICE_NOT_REGISTERED) {
      // 패스코드 요청
      await authClient.requestPasscode({ email: KAKAO_EMAIL, password: KAKAO_PASSWORD });
      console.error(
        '[auth] 기기 미등록 상태입니다.\n' +
        '카카오 계정 이메일로 패스코드가 발송되었습니다.\n' +
        '환경변수 KAKAO_PASSCODE=<패스코드> 를 설정 후 재실행하세요.',
      );
      process.exit(1);
    }
    console.error('[auth] Login failed detail:', JSON.stringify(loginResult));
    throw new Error(`[auth] Login failed: status=${loginResult.status}`);
  }

  const credential: OAuthCredential = loginResult.result;
  saveCredential(credential);
  console.log('[auth] Login successful');
  return credential;
}

// ── Loco 연결 + 재연결 ─────────────────────────────────────
async function connectWithRetry(credential: OAuthCredential): Promise<void> {
  let attempt = 0;

  while (true) {
    try {
      await connect(credential);
      // connect()가 정상 종료되면 재연결
      console.warn('[client] Session ended, reconnecting...');
    } catch (err) {
      console.error(`[client] Connection error (attempt ${attempt + 1}):`, err);
    }

    attempt++;
    const delay = Math.min(RECONNECT_BASE_MS * 2 ** attempt, RECONNECT_MAX_MS);
    console.log(`[client] Reconnecting in ${delay / 1000}s...`);
    await sleep(delay);
  }
}

async function connect(credential: OAuthCredential): Promise<void> {
  const client = new TalkClient();

  client.on('chat', (data, channel) => {
    handleChat(data, channel, client).catch((err) =>
      console.error('[bot] Unhandled error in handleChat:', err),
    );
  });

  client.on('disconnected', (reason) => {
    console.warn(`[client] Disconnected: ${reason}`);
  });

  const loginResult = await client.login(credential);
  if (!loginResult.success) {
    throw new Error(`[client] TalkClient login failed: status=${loginResult.status}`);
  }

  console.log(`[client] Connected as ${client.clientUser.userId}`);

  // 연결 유지 — 'disconnected' 이벤트가 올 때까지 대기
  await new Promise<void>((resolve) => {
    client.once('disconnected', () => resolve());
  });

  client.close();
}

// ── 크리덴셜 영속화 ────────────────────────────────────────
function loadCredential(): OAuthCredential | null {
  try {
    const raw = fs.readFileSync(CREDENTIAL_PATH, 'utf-8');
    return JSON.parse(raw) as OAuthCredential;
  } catch {
    return null;
  }
}

function saveCredential(credential: OAuthCredential): void {
  try {
    fs.mkdirSync(path.dirname(CREDENTIAL_PATH), { recursive: true });
    fs.writeFileSync(CREDENTIAL_PATH, JSON.stringify(credential), { mode: 0o600 });
    console.log(`[auth] Credential saved to ${CREDENTIAL_PATH}`);
  } catch (err) {
    console.warn('[auth] Failed to save credential:', err);
  }
}

// ── 유틸 ──────────────────────────────────────────────────
function requireEnv(key: string): string {
  const val = process.env[key];
  if (!val) throw new Error(`Environment variable ${key} is required`);
  return val;
}

/** 이메일에서 결정론적 UUID 생성 (기기 재등록 방지) */
function deriveUUID(email: string): string {
  let hash = 0;
  for (let i = 0; i < email.length; i++) {
    hash = (Math.imul(31, hash) + email.charCodeAt(i)) | 0;
  }
  const h = Math.abs(hash).toString(16).padStart(8, '0');
  return `yangobot-${h}-0000-0000-${h}${h}`.slice(0, 36);
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
