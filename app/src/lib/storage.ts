import { Preferences } from '@capacitor/preferences';

const KEYS = {
  serverUrl: 'server_url',
  defaultPublic: 'default_public',
  defaultExpire: 'default_expire',
} as const;

export async function getServerUrl(): Promise<string | null> {
  const { value } = await Preferences.get({ key: KEYS.serverUrl });
  return value;
}

export async function setServerUrl(url: string): Promise<void> {
  await Preferences.set({ key: KEYS.serverUrl, value: url.replace(/\/+$/, '') });
}

export async function getDefaultPublic(): Promise<boolean> {
  const { value } = await Preferences.get({ key: KEYS.defaultPublic });
  return value !== 'false';
}

export async function setDefaultPublic(v: boolean): Promise<void> {
  await Preferences.set({ key: KEYS.defaultPublic, value: String(v) });
}

export async function getDefaultExpire(): Promise<string> {
  const { value } = await Preferences.get({ key: KEYS.defaultExpire });
  return value || '3mo';
}

export async function setDefaultExpire(v: string): Promise<void> {
  await Preferences.set({ key: KEYS.defaultExpire, value: v });
}

export const EXPIRE_OPTIONS = [
  { value: 'never', label: '永不过期' },
  { value: '1d', label: '1 天' },
  { value: '7d', label: '7 天' },
  { value: '30d', label: '30 天' },
  { value: '1mo', label: '1 个月' },
  { value: '3mo', label: '3 个月' },
  { value: '1y', label: '1 年' },
] as const;
