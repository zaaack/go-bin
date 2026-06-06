import type { CapacitorConfig } from '@capacitor/cli';

const config: CapacitorConfig = {
  appId: 'com.gobin.app',
  appName: 'Go Bin',
  webDir: 'dist',
  server: {
    androidScheme: 'http',
  },
};

export default config;
