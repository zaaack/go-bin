import { useEffect, useState } from 'react';
import {
  getServerUrl,
  setServerUrl,
  getDefaultPublic,
  setDefaultPublic,
  getDefaultExpire,
  setDefaultExpire,
  EXPIRE_OPTIONS,
} from '../lib/storage';
import { testConnection } from '../lib/api';

interface Props {
  onBack: () => void;
  onUrlChange: (url: string) => void;
}

export default function Settings({ onBack, onUrlChange }: Props) {
  const [url, setUrl] = useState('');
  const [isPublic, setIsPublic] = useState(false);
  const [expire, setExpire] = useState('3mo');
  const [testing, setTesting] = useState(false);
  const [urlError, setUrlError] = useState('');
  const [urlSuccess, setUrlSuccess] = useState(false);

  useEffect(() => {
    (async () => {
      const u = await getServerUrl();
      if (u) setUrl(u);
      setIsPublic(await getDefaultPublic());
      setExpire(await getDefaultExpire());
    })();
  }, []);

  async function handleSaveUrl(e: React.FormEvent) {
    e.preventDefault();
    const trimmed = url.trim().replace(/\/+$/, '');
    if (!trimmed) {
      setUrlError('请输入服务地址');
      return;
    }
    setTesting(true);
    setUrlError('');
    setUrlSuccess(false);

    const ok = await testConnection(trimmed);
    setTesting(false);

    if (!ok) {
      setUrlError('无法连接到服务');
      return;
    }

    await setServerUrl(trimmed);
    setUrl(trimmed);
    setUrlSuccess(true);
    onUrlChange(trimmed);
    setTimeout(() => setUrlSuccess(false), 2000);
  }

  async function togglePublic() {
    const next = !isPublic;
    setIsPublic(next);
    await setDefaultPublic(next);
  }

  async function handleExpireChange(val: string) {
    setExpire(val);
    await setDefaultExpire(val);
  }

  return (
    <div className="settings-page">
      <nav className="top-bar">
        <button className="icon-btn" onClick={onBack} title="返回">
          ←
        </button>
        <span className="title">设置</span>
        <span className="icon-btn" />
      </nav>

      <div className="settings-content">
        <section>
          <h2>服务地址</h2>
          <form onSubmit={handleSaveUrl} className="url-form">
            <input
              type="url"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="http://192.168.1.100:8080"
            />
            <button type="submit" disabled={testing}>
              {testing ? '...' : '保存'}
            </button>
          </form>
          {urlError && <p className="error">{urlError}</p>}
          {urlSuccess && <p className="success">已保存</p>}
        </section>

        <section>
          <h2>分享默认设置</h2>
          <div className="setting-row">
            <span>默认公开</span>
            <label className="toggle">
              <input
                type="checkbox"
                checked={isPublic}
                onChange={togglePublic}
              />
              <span className="toggle-slider" />
            </label>
          </div>
          <div className="setting-row">
            <span>过期时间</span>
            <select
              value={expire}
              onChange={(e) => handleExpireChange(e.target.value)}
            >
              {EXPIRE_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>
        </section>
      </div>
    </div>
  );
}
