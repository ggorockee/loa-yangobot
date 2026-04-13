/**
 * yangobot REST API 호출 모듈
 * 클러스터 내부에서는 http://yangobot, 로컬에서는 http://localhost:8080
 */

const BASE_URL = process.env.YANGOBOT_API_URL ?? 'http://localhost:8080';

interface BotResponse {
  text: string;
}

async function callAPI(path: string): Promise<string> {
  const url = `${BASE_URL}/api/v1/${path}`;
  const res = await fetch(url, { signal: AbortSignal.timeout(10_000) });

  if (!res.ok) {
    const body = await res.text().catch(() => '');
    throw new Error(`API ${res.status}: ${body}`);
  }

  const json = (await res.json()) as BotResponse;
  return json.text ?? '';
}

export async function getArmory(name: string): Promise<string> {
  return callAPI(`armory/${encodeURIComponent(name)}`);
}

export async function getLopec(name: string): Promise<string> {
  return callAPI(`lopec/${encodeURIComponent(name)}`);
}

export async function getCharacter(name: string): Promise<string> {
  return callAPI(`character/${encodeURIComponent(name)}`);
}

export async function getExpedition(name: string): Promise<string> {
  return callAPI(`expedition/${encodeURIComponent(name)}`);
}

export async function getDistribute(n: 4 | 8, query: string): Promise<string> {
  const q = query.replace(/,/g, '');
  return callAPI(`distribute/${n}/${encodeURIComponent(q)}`);
}
