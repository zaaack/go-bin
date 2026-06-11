import { App as CapApp } from '@capacitor/app';
import { getServerUrl, getDefaultPublic, getDefaultExpire } from './storage';
import { uploadText, uploadFile } from './api';

export interface SharedData {
  text?: string;
  files?: File[];
}

export interface UploadResult {
  url: string;
}

/**
 * Called from native MainActivity when a share intent is received.
 * Exposed on window.__handleShareIntent for the native side to call.
 */
export async function handleShareIntent(
  sharedText?: string,
  sharedFiles?: Array<{ name: string; uri: string; dataUrl: string }>,
  nativeResultUrl?: string,
): Promise<UploadResult | null> {
  if (nativeResultUrl) {
    return { url: nativeResultUrl };
  }

  const baseUrl = await getServerUrl();
  if (!baseUrl) return null;

  const isPublic = await getDefaultPublic();
  const expire = await getDefaultExpire();

  if (sharedFiles && sharedFiles.length > 0) {
    const files: File[] = [];
    for (const f of sharedFiles) {
      if (f.dataUrl) {
        const blob = dataURLtoBlob(f.dataUrl);
        files.push(new File([blob], f.name, { type: blob.type }));
      }
    }
    if (files.length > 0) {
      const url = await uploadFile(baseUrl, files[0], { isPublic, expire });
      return { url };
    }
  }

  if (sharedText) {
    const url = await uploadText(baseUrl, sharedText, { isPublic, expire });
    return { url };
  }

  return null;
}

function dataURLtoBlob(dataUrl: string): Blob {
  const [header, base64] = dataUrl.split(',');
  const mime = header.match(/:(.*?);/)?.[1] || 'application/octet-stream';
  const bytes = atob(base64);
  const arr = new Uint8Array(bytes.length);
  for (let i = 0; i < bytes.length; i++) arr[i] = bytes.charCodeAt(i);
  return new Blob([arr], { type: mime });
}

// Register global handler for native share intent
declare global {
  interface Window {
    __handleShareIntent?: (
      text?: string,
      files?: Array<{ name: string; uri: string; dataUrl: string }>,
      nativeResultUrl?: string,
    ) => void;
    __handleShareError?: (error: string) => void;
    __shareResult?: UploadResult | null;
    __shareError?: string;
  }
}

export function registerShareHandler(
  onResult: (result: UploadResult | null, error?: string) => void,
) {
  window.__handleShareIntent = async (text, files, nativeResultUrl) => {
    try {
      const result = await handleShareIntent(text, files, nativeResultUrl);
      window.__shareResult = result;
      onResult(result);
    } catch (err: any) {
      window.__shareError = err.message;
      onResult(null, err.message);
    }
  };

  window.__handleShareError = (error: string) => {
    window.__shareError = error;
    onResult(null, error);
  };
}

// Listen for app state changes to detect share intent on Android
export function setupShareListener(
  onResult: (result: UploadResult | null, error?: string) => void,
) {
  CapApp.addListener('appStateChange', async ({ isActive }) => {
    if (isActive) {
      // Check if there's a pending share result from native
      if (window.__shareResult) {
        onResult(window.__shareResult);
        window.__shareResult = null;
      }
    }
  });
}
