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
    redirect: 'manual',
  });

  // Server returns 303 redirect to /s/{slug}
  if (resp.status === 303 || resp.status === 302) {
    const location = resp.headers.get('Location');
    if (location) {
      const slug = location.replace(/^.*\/s\//, '');
      return `${baseUrl}/s/${slug}`;
    }
  }

  if (resp.ok) {
    // If redirect is followed, try to extract slug from final URL
    const url = new URL(resp.url);
    if (url.pathname.startsWith('/s/')) {
      return resp.url;
    }
  }

  const text = await resp.text().catch(() => 'Unknown error');
  throw new Error(`Upload failed (${resp.status}): ${text}`);
}
