import { useEffect, useState } from 'react';
import { App as CapApp } from '@capacitor/app';
import { getServerUrl } from './lib/storage';
import Setup from './pages/Setup';
import Home from './pages/Home';
import Settings from './pages/Settings';
import './index.css';

type Page = 'setup' | 'home' | 'settings';

function App() {
  const [page, setPage] = useState<Page>('setup');
  const [serverUrl, setServerUrl] = useState('');

  useEffect(() => {
    getServerUrl().then((url) => {
      if (url) {
        setServerUrl(url);
        setPage('home');
      }
    });
  }, []);

  useEffect(() => {
    const handler = CapApp.addListener('backButton', ({ canGoBack }) => {
      if (canGoBack) {
        window.history.back();
      } else {
        CapApp.exitApp();
      }
    });
    return () => { handler.then((h) => h.remove()); };
  }, []);

  if (page === 'setup') {
    return (
      <Setup
        onDone={(url) => {
          setServerUrl(url);
          setPage('home');
        }}
      />
    );
  }

  if (page === 'settings') {
    return (
      <Settings
        onBack={() => setPage('home')}
        onUrlChange={(url) => setServerUrl(url)}
      />
    );
  }

  return (
    <Home
      serverUrl={serverUrl}
      onOpenSettings={() => setPage('settings')}
    />
  );
}

export default App;
