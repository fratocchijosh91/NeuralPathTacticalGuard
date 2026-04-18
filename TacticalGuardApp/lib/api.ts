import { DEFAULT_RAILWAY_BASE } from "../constants/defaults";

export type DetectedDevice = {
  id: string;
  type: string;
  online: boolean;
  starlink_ms?: number;
  device_ms?: number;
  last_seen: string;
};

type DevicesPayload = {
  devices?: DetectedDevice[];
};

function joinUrl(base: string, path: string): string {
  const b = base.replace(/\/+$/, "");
  const p = path.startsWith("/") ? path : `/${path}`;
  return `${b}${p}`;
}

export async function fetchServerConnected(baseUrl: string): Promise<boolean> {
  try {
    const res = await fetch(joinUrl(baseUrl, "/healthz"), {
      method: "GET",
      headers: { Accept: "application/json" },
    });
    if (!res.ok) return false;
    const body = (await res.json()) as { status?: string };
    return body.status === "ok";
  } catch {
    return false;
  }
}

export async function fetchDetectedDevices(baseUrl: string): Promise<DetectedDevice[]> {
  const res = await fetch(joinUrl(baseUrl, "/v1/detected-devices"), {
    method: "GET",
    headers: { Accept: "application/json" },
  });
  if (res.status === 404) {
    return [];
  }
  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`);
  }
  const json = (await res.json()) as DevicesPayload | DetectedDevice[];
  if (Array.isArray(json)) {
    return json;
  }
  return json.devices ?? [];
}

export function getRailwayBaseUrl(): string {
  const fromEnv = process.env.EXPO_PUBLIC_RAILWAY_BASE_URL;
  if (fromEnv && fromEnv.trim().length > 0) {
    return fromEnv.trim().replace(/\/+$/, "");
  }
  return DEFAULT_RAILWAY_BASE;
}
