var API_BASE = "https://yangobot.ggorockee.com/api/v1";

// Utils.getWebText()가 HTML로 감싸서 반환하므로 body 텍스트만 추출 후 JSON 파싱
function parseResponse(html) {
  if (!html) return null;
  var body = html.match(/<body[^>]*>([\s\S]*?)<\/body>/i);
  var text = body ? body[1] : html;
  text = text.replace(/<[^>]*>/g, "").trim();
  try {
    var obj = JSON.parse(text);
    return obj.text || text;
  } catch (e) {
    return text || null;
  }
}

function response(room, msg, sender, isGroupChat, replier, imageDB, packageName) {
  var res, m;

  m = msg.match(/^[.·]군장\s+(.+)/);
  if (m) {
    res = parseResponse(Utils.getWebText(API_BASE + "/armory/" + encodeURIComponent(m[1].trim())));
    replier.reply(res || "캐릭터를 찾을 수 없습니다.");
    return;
  }

  m = msg.match(/^[.·]스펙\s+(.+)/);
  if (m) {
    res = parseResponse(Utils.getWebText(API_BASE + "/lopec/" + encodeURIComponent(m[1].trim())));
    replier.reply(res || "캐릭터를 찾을 수 없습니다.");
    return;
  }

  m = msg.match(/^[.·]캐릭터\s+(.+)/);
  if (m) {
    res = parseResponse(Utils.getWebText(API_BASE + "/character/" + encodeURIComponent(m[1].trim())));
    replier.reply(res || "캐릭터를 찾을 수 없습니다.");
    return;
  }

  m = msg.match(/^[.·]원정대\s+(.+)/);
  if (m) {
    res = parseResponse(Utils.getWebText(API_BASE + "/expedition/" + encodeURIComponent(m[1].trim())));
    replier.reply(res || "캐릭터를 찾을 수 없습니다.");
    return;
  }

  // .ㄱㅁ8 49000 또는 .ㄱㅁ8 아드레날린 — 분배금 계산 (숫자: 직접 입력, 텍스트: 거래소 시세 조회)
  m = msg.match(/^[.·]ㄱㅁ([48])\s+(.+)/);
  if (m) {
    var n = m[1];
    var query = m[2].trim().replace(/,/g, "");
    res = parseResponse(Utils.getWebText(API_BASE + "/distribute/" + n + "/" + encodeURIComponent(query)));
    replier.reply(res || "계산 중 오류가 발생했습니다.");
    return;
  }
}

function onCreate(savedInstanceState, activity) {}
function onResume(activity) {}
function onPause(activity) {}
function onStop(activity) {}
function onDestroy(activity) {}
function onBackPressed(activity) {}
function onActivityResult(requestCode, resultCode, data, activity) {}
function onStartCompile(activity) {}
