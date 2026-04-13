/**
 * 메시지 파싱 및 명령어 처리
 * messagerbot.js의 정규식 패턴을 그대로 포팅
 *
 * 관리자 전용 명령어:
 *   .방추가 <오픈채팅 링크>  — 실행 중인 봇이 세션 끊김 없이 방에 참여
 */

import { TalkChatData, TalkChannel, TalkClient } from 'node-kakao';
import * as api from './api';

// 관리자 카카오 userId (bson Long.toString())
// 쉼표 구분, 예: ADMIN_USER_IDS=123456789,987654321
const ADMIN_IDS = new Set(
  (process.env.ADMIN_USER_IDS ?? '').split(',').map((s) => s.trim()).filter(Boolean),
);

export async function handleChat(
  data: TalkChatData,
  channel: TalkChannel,
  client: TalkClient,
): Promise<void> {
  const text = data.text.trim();
  if (!text) return;

  const senderId = data.chat.sender.userId.toString();
  // 관리자 설정 전 userId 확인용 (ADMIN_USER_IDS 미설정 시에만 출력)
  if (ADMIN_IDS.size === 0 && text.startsWith('.')) {
    console.log(`[bot] senderId=${senderId} msg="${text}"`);
  }
  let reply: string | null = null;

  try {
    reply = await dispatch(text, senderId, client);
  } catch (err) {
    console.error(`[bot] dispatch error (msg="${text}"):`, err);
    reply = '처리 중 오류가 발생했습니다.';
  }

  if (reply == null) return;

  const result = await channel.sendChat(reply);
  if (!result.success) {
    console.error(`[bot] sendChat failed: status=${result.status}`);
  }
}

async function dispatch(
  text: string,
  senderId: string,
  client: TalkClient,
): Promise<string | null> {
  let m: RegExpMatchArray | null;

  // ── 관리자 전용 ──────────────────────────────────────────
  // .방추가 <오픈채팅 URL> [비밀번호]
  m = text.match(/^[.·]방추가\s+(https?:\/\/open\.kakao\.com\/\S+)(?:\s+(\S+))?/);
  if (m) {
    if (!ADMIN_IDS.has(senderId)) return '권한이 없습니다.';
    return joinRoom(client, m[1].trim(), m[2]?.trim());
  }

  // ── 일반 명령어 ──────────────────────────────────────────
  // .군장 <닉네임>
  m = text.match(/^[.·]군장\s+(.+)/);
  if (m) return api.getArmory(m[1].trim());

  // .스펙 <닉네임>
  m = text.match(/^[.·]스펙\s+(.+)/);
  if (m) return api.getLopec(m[1].trim());

  // .캐릭터 <닉네임>
  m = text.match(/^[.·]캐릭터\s+(.+)/);
  if (m) return api.getCharacter(m[1].trim());

  // .원정대 <닉네임>
  m = text.match(/^[.·]원정대\s+(.+)/);
  if (m) return api.getExpedition(m[1].trim());

  // .ㄱㅁ8 <금액 or 각인서명>  /  .ㄱㅁ4 <금액 or 각인서명>
  m = text.match(/^[.·]ㄱㅁ([48])\s+(.+)/);
  if (m) {
    const n = parseInt(m[1], 10) as 4 | 8;
    const query = m[2].trim();
    return api.getDistribute(n, query);
  }

  return null;
}

// ── 방 참여 ─────────────────────────────────────────────────
async function joinRoom(client: TalkClient, linkURL: string, passcode?: string): Promise<string> {
  // 1) 링크 정보 조회 (linkId 획득)
  const infoResult = await client.channelList.open.getJoinInfo(linkURL);
  if (!infoResult.success) {
    return `방 정보 조회 실패 (status=${infoResult.status})`;
  }

  const link = infoResult.result.openLink;

  // 이미 참여 중인지 확인
  const existing = client.channelList.open.getChannelByLinkId(link.linkId);
  if (existing) {
    return `이미 참여 중인 방입니다.`;
  }

  // 2) 메인 프로필로 참여 ({} = OpenLinkMainProfile)
  const joinResult = await client.channelList.open.joinChannel(link, {}, passcode);
  if (!joinResult.success) {
    return `방 참여 실패 (status=${joinResult.status})`;
  }

  console.log(`[bot] Joined open channel: ${link.linkName} (linkId=${link.linkId})`);
  return `"${link.linkName}" 방 참여 완료`;
}
