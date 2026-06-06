export interface UploadOptions {
  isPublic: boolean;
  expire: string;
}

export async function testConnection(baseUrl: string): Promise<boolean> {
  try {
    const resp = await fetch(baseUrl, { method: 'GET', mode: 'no-cors' });
    // no-cors returns opaque response, status 0 means it reached the server
    return resp.type === 'opaque' || resp.ok;
  } catch {
    return false;
  }
}

export async function uploadText(
  baseUrl: string,
  text: string,
  opts: UploadOptions,
): Promise<string> {
  const form = new FormData();
  form.append('kind', 'text');
  form.append('text', text);
  form.append('is_public', opts.isPublic ? 'on' : 'off');
  form.append('expire', opts.expire);
  return doUpload(baseUrl, form);
}

export async function uploadFile(
  baseUrl: string,
  file: File,
  opts: UploadOptions,
): Promise<string> {
  const form = new FormData();
  form.append('kind', 'file');
  form.append('files', file);
  form.append('is_public', opts.isPublic ? 'on' : 'off');
  form.append('expire', opts.expire);
  return doUpload(baseUrl, form);
}

async function doUpload(baseUrl: string, form: FormData): Promise<string> {
  const resp = await fetch(`${baseUrl}/shares`, {
    method: 'POST',
    body: form,
  });

  if (!resp.ok) {
    const text = await resp.text().catch(() => 'Unknown error');
    throw new Error(`Upload failed (${resp.status}): ${text}`);
  }

  // After following redirect, final URL is /s/{slug}
  const finalUrl = resp.url;
  if (finalUrl.includes('/s/')) {
    return finalUrl;
  }

  // Fallback: if redirect wasn't followed, return base URL
  return baseUrl;
}
