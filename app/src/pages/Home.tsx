import { useEffect, useRef, useState } from 'react';
import { registerShareHandler, type UploadResult } from '../lib/share-handler';

interface Props {
  serverUrl: string;
  onOpenSettings: () => void;
}

export default function Home({ serverUrl, onOpenSettings }: Props) {
  const iframeRef = useRef<HTMLIFrameElement>(null);
  const [shareResult, setShareResult] = useState<UploadResult | null>(null);
  const [shareError, setShareError] = useState<string | null>(null);
  const [showToast, setShowToast] = useState(false);
  const [nativeLoading, setNativeLoading] = useState(false);
  const [loadingFileName, setLoadingFileName] = useState<string | null>(null);

  useEffect(() => {
    registerShareHandler(
      (result, error) => {
        if (result) {
          setShareResult(result);
          setShareError(null);
          setShowToast(true);
          setTimeout(() => setShowToast(false), 5000);
        } else if (error) {
          setShareResult(null);
          setShareError(error);
          setShowToast(true);
          setTimeout(() => setShowToast(false), 5000);
        }
      },
      (fileName) => {
        setNativeLoading(true);
        setLoadingFileName(fileName || null);
      },
      () => {
        setNativeLoading(false);
        setLoadingFileName(null);
      },
    );
  }, []);

  function handleRefresh() {
    if (iframeRef.current) {
      iframeRef.current.src = serverUrl;
    }
  }

  function dismissToast() {
    setShowToast(false);
  }

  return (
    <div className="home-page">
      <nav className="top-bar">
        <button className="icon-btn" onClick={handleRefresh} title="刷新">
          ↻
        </button>
        <span className="title">Go Bin</span>
        <button className="icon-btn" onClick={onOpenSettings} title="设置">
          ⚙
        </button>
      </nav>

      <iframe
        ref={iframeRef}
        src={serverUrl}
        className="web-frame"
        title="Go Bin"
      />

      {nativeLoading && (
        <div className="loading-overlay">
          <div className="loading-card">
            <div className="loading-spinner" />
            <p className="loading-text">正在上传…</p>
            {loadingFileName && (
              <p className="loading-filename">{loadingFileName}</p>
            )}
          </div>
        </div>
      )}

      {showToast && (
        <div className={`toast ${shareError ? 'toast-error' : 'toast-success'}`}>
          <div className="toast-content">
            {shareError ? (
              <p>分享失败: {shareError}</p>
            ) : (
              <>
                <p>分享成功!</p>
                {shareResult && (
                  <a
                    href={shareResult.url}
                    target="_blank"
                    rel="noopener"
                    className="toast-link"
                  >
                    {shareResult.url.replace(serverUrl, '')}
                  </a>
                )}
              </>
            )}
          </div>
          <button className="toast-close" onClick={dismissToast}>✕</button>
        </div>
      )}
    </div>
  );
}
