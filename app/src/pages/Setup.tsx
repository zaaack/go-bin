import { useState } from 'react';
import { setServerUrl } from '../lib/storage';
import { testConnection } from '../lib/api';

interface Props {
  onDone: (url: string) => void;
}

export default function Setup({ onDone }: Props) {
  const [url, setUrl] = useState('');
  const [testing, setTesting] = useState(false);
  const [error, setError] = useState('');

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const trimmed = url.trim().replace(/\/+$/, '');
    if (!trimmed) {
      setError('请输入服务地址');
      return;
    }
    setTesting(true);
    setError('');

    const ok = await testConnection(trimmed);
    setTesting(false);

    if (!ok) {
      setError('无法连接到服务，请检查地址');
      return;
    }

    await setServerUrl(trimmed);
    onDone(trimmed);
  }

  return (
    <div className="setup-page">
      <div className="setup-card">
        <h1>Go Bin</h1>
        <p className="subtitle">文件 / 文本 / 链接分享服务</p>
        <form onSubmit={handleSubmit}>
          <label htmlFor="server-url">服务地址</label>
          <input
            id="server-url"
            type="url"
            placeholder="http://192.168.1.100:8080"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            autoFocus
          />
          {error && <p className="error">{error}</p>}
          <button type="submit" disabled={testing}>
            {testing ? '连接中...' : '连接'}
          </button>
        </form>
      </div>
    </div>
  );
}
