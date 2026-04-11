/**
 * yangobot 메신저봇R 스크립트
 *
 * [설치 방법]
 * 1. Android 기기에 "메신저봇R" 앱 설치
 * 2. 봇 전용 카카오 계정으로 로그인 후 목표 오픈채팅방에 참여
 * 3. 메신저봇R → 새 봇 생성 → 이 파일 내용을 스크립트에 붙여넣기
 * 4. TARGET_ROOMS에 실제 방 이름 입력 후 봇 활성화
 * 5. 알림 접근 권한 허용
 *
 * [지원 명령어]
 *   .군장 <캐릭터명>    — 각인·카드·보석·아크그리드 포함 군장 정보
 *   .스펙 <캐릭터명>    — 로펙 스펙 점수 및 티어
 *   .캐릭터 <캐릭터명>  — 서버·직업·아이템 레벨 기본 정보
 *
 * [도트 접두사]
 *   . (마침표) 또는 · (가운뎃점) 모두 인식
 */

// ─── 설정 ──────────────────────────────────────────────────────────────────────

/** 봇이 응답할 오픈채팅방 이름 목록. 비워두면 모든 방에서 응답합니다. */
var TARGET_ROOMS = [
  // "로아 길드방",
  // "로스트아크 공략방",
];

/** yangobot API 서버 주소 */
var API_BASE = "https://yangobot.ggorockee.com/api/v1";

/** HTTP 요청 타임아웃 (밀리초) */
var TIMEOUT_MS = 10000;

// ─── 명령어 정의 ────────────────────────────────────────────────────────────────

var COMMANDS = [
  { pattern: /^[.·]군장\s+(.+)$/,    resource: "armory"    },
  { pattern: /^[.·]스펙\s+(.+)$/,    resource: "lopec"     },
  { pattern: /^[.·]캐릭터\s+(.+)$/,  resource: "character" },
];

// ─── 메인 응답 함수 ─────────────────────────────────────────────────────────────

function response(room, msg, sender, isGroupChat, replier, imageDB, packageName) {
  // 그룹채팅만 응답
  if (!isGroupChat) return;

  // 지정된 방만 응답 (TARGET_ROOMS가 비어 있으면 전체 허용)
  if (TARGET_ROOMS.length > 0 && TARGET_ROOMS.indexOf(room) === -1) return;

  var matched = null;
  for (var i = 0; i < COMMANDS.length; i++) {
    var m = msg.match(COMMANDS[i].pattern);
    if (m) {
      matched = { resource: COMMANDS[i].resource, name: m[1].trim() };
      break;
    }
  }
  if (!matched) return;

  var url = API_BASE + "/" + matched.resource + "/" + encodeURIComponent(matched.name);

  try {
    var result = fetchText(url);
    if (result && result.trim().length > 0) {
      replier.reply(result.trim());
    } else {
      replier.reply("캐릭터를 찾을 수 없습니다: " + matched.name);
    }
  } catch (e) {
    replier.reply("오류가 발생했습니다. 잠시 후 다시 시도해 주세요.\n(" + e.message + ")");
  }
}

// ─── HTTP GET 헬퍼 ──────────────────────────────────────────────────────────────

/**
 * URL에서 텍스트를 가져옵니다.
 * 메신저봇R API인 Utils.getWebText()를 사용합니다.
 * 응답 코드가 200이 아니면 예외를 던집니다.
 */
function fetchText(url) {
  // Utils.getWebText는 성공 시 문자열, 실패 시 null/빈 문자열을 반환합니다.
  var text = Utils.getWebText(url);
  if (text === null || text === undefined) {
    throw new Error("API 응답 없음: " + url);
  }
  return text;
}

// ─── 봇 시작/종료 이벤트 (선택) ────────────────────────────────────────────────

function onCreate(savedInstanceState, activity) {
  // 필요 시 초기화 로직 추가
}

function onStartCompile() {
  // 컴파일 시작 시 호출
}
