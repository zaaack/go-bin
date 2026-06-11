package com.gobin.app;

import android.content.Intent;
import android.content.SharedPreferences;
import android.net.Uri;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.webkit.MimeTypeMap;

import com.getcapacitor.BridgeActivity;

import java.io.InputStream;
import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.URL;
import java.util.ArrayList;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

public class MainActivity extends BridgeActivity {

    private static final String PREFS_NAME = "CapacitorStorage";
    private final ExecutorService executor = Executors.newSingleThreadExecutor();
    private final Handler mainHandler = new Handler(Looper.getMainLooper());

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        handleShareIntent(getIntent());
    }

    @Override
    protected void onNewIntent(Intent intent) {
        super.onNewIntent(intent);
        setIntent(intent);
        handleShareIntent(intent);
    }

    private void handleShareIntent(Intent intent) {
        String action = intent.getAction();
        String type = intent.getType();
        if (action == null || type == null) return;
        if (!Intent.ACTION_SEND.equals(action) && !Intent.ACTION_SEND_MULTIPLE.equals(action)) return;

        if (Intent.ACTION_SEND.equals(action)) {
            handleSendSingle(intent, type);
        } else {
            handleSendMultiple(intent);
        }
    }

    private void handleSendSingle(Intent intent, String type) {
        Uri fileUri = intent.getParcelableExtra(Intent.EXTRA_STREAM);
        if (fileUri != null) {
            String name = getFileName(fileUri);
            String mimeType = resolveMimeType(fileUri, name);
            String serverUrl = getPref("server_url");
            if (serverUrl == null) return;

            executor.execute(() -> {
                mainHandler.post(() -> evalJs("if(window.__showNativeLoading){window.__showNativeLoading(" + jsonString(name) + ");}"));
                String resultUrl = uploadFileNative(fileUri, name, mimeType, serverUrl);
                mainHandler.post(() -> {
                    evalJs("if(window.__hideNativeLoading){window.__hideNativeLoading();}");
                    notifyJs(resultUrl, resultUrl == null ? "上传失败" : null);
                });
            });
            return;
        }

        String sharedText = intent.getStringExtra(Intent.EXTRA_TEXT);
        if (sharedText != null) {
            pollAndExecute("if(window.__handleShareIntent){window.__handleShareIntent("
                    + jsonString(sharedText) + ");}");
        }
    }

    private void handleSendMultiple(Intent intent) {
        ArrayList<Uri> uris = intent.getParcelableArrayListExtra(Intent.EXTRA_STREAM);
        if (uris == null || uris.isEmpty()) return;

        String serverUrl = getPref("server_url");
        if (serverUrl == null) return;

        executor.execute(() -> {
            mainHandler.post(() -> evalJs("if(window.__showNativeLoading){window.__showNativeLoading();}"));
            String finalUrl = null;
            for (Uri uri : uris) {
                String name = getFileName(uri);
                String mimeType = resolveMimeType(uri, name);
                String url = uploadFileNative(uri, name, mimeType, serverUrl);
                if (url != null) {
                    finalUrl = url;
                    break;
                }
            }
            String captured = finalUrl;
            mainHandler.post(() -> {
                evalJs("if(window.__hideNativeLoading){window.__hideNativeLoading();}");
                notifyJs(captured, captured == null ? "上传失败" : null);
            });
        });
    }

    private void notifyJs(String resultUrl, String error) {
        String js;
        if (resultUrl != null) {
            js = "if(window.__handleShareIntent){window.__handleShareIntent(null,[],"
                    + jsonString(resultUrl) + ");}";
        } else if (error != null) {
            js = "if(window.__handleShareError){window.__handleShareError("
                    + jsonString(error) + ");}";
        } else {
            return;
        }
        pollAndExecute(js);
    }

    private String uploadFileNative(Uri fileUri, String fileName, String mimeType, String serverUrl) {
        InputStream is = null;
        OutputStream os = null;
        HttpURLConnection conn = null;
        try {
            String boundary = "----GoBin" + System.currentTimeMillis();
            URL url = new URL(serverUrl + "/shares");
            conn = (HttpURLConnection) url.openConnection();
            conn.setRequestMethod("POST");
            conn.setDoOutput(true);
            conn.setInstanceFollowRedirects(false);
            conn.setConnectTimeout(15000);
            conn.setReadTimeout(60000);
            conn.setRequestProperty("Content-Type", "multipart/form-data; boundary=" + boundary);

            os = conn.getOutputStream();

            writeField(os, boundary, "kind", "file");
            writeField(os, boundary, "is_public", getDefault("default_public", "true").equals("true") ? "on" : "off");
            writeField(os, boundary, "expire", getDefault("default_expire", "3mo"));

            byte[] lineEnd = "\r\n".getBytes();
            os.write(("--" + boundary + "\r\n").getBytes());
            os.write(("Content-Disposition: form-data; name=\"files\"; filename=\"" + fileName + "\"\r\n").getBytes());
            os.write(("Content-Type: " + mimeType + "\r\n").getBytes());
            os.write(lineEnd);

            is = getContentResolver().openInputStream(fileUri);
            if (is == null) return null;
            byte[] buf = new byte[8192];
            int n;
            while ((n = is.read(buf)) != -1) {
                os.write(buf, 0, n);
            }

            os.write(lineEnd);
            os.write(("--" + boundary + "--\r\n").getBytes());
            os.flush();

            int code = conn.getResponseCode();
            if (code == 303) {
                String location = conn.getHeaderField("Location");
                if (location != null && location.startsWith("/")) {
                    return serverUrl + location;
                }
                return location;
            }
            return null;
        } catch (Exception e) {
            return null;
        } finally {
            try { if (is != null) is.close(); } catch (Exception ignored) {}
            try { if (os != null) os.close(); } catch (Exception ignored) {}
            if (conn != null) conn.disconnect();
        }
    }

    private void writeField(OutputStream os, String boundary, String name, String value) throws Exception {
        byte[] lineEnd = "\r\n".getBytes();
        os.write(("--" + boundary + "\r\n").getBytes());
        os.write(("Content-Disposition: form-data; name=\"" + name + "\"\r\n").getBytes());
        os.write(lineEnd);
        os.write(value.getBytes());
        os.write(lineEnd);
    }

    private String getPref(String key) {
        SharedPreferences prefs = getSharedPreferences(PREFS_NAME, MODE_PRIVATE);
        return prefs.getString(key, null);
    }

    private String getDefault(String key, String fallback) {
        String val = getPref(key);
        return val != null ? val : fallback;
    }

    private String resolveMimeType(Uri uri, String name) {
        String mimeType = getContentResolver().getType(uri);
        if (mimeType == null && name != null) {
            int dot = name.lastIndexOf('.');
            if (dot >= 0) {
                mimeType = MimeTypeMap.getSingleton().getMimeTypeFromExtension(
                        name.substring(dot + 1));
            }
        }
        return mimeType != null ? mimeType : "application/octet-stream";
    }

    private String getFileName(Uri uri) {
        String name = null;
        try (android.database.Cursor c = getContentResolver().query(uri,
                new String[]{android.provider.OpenableColumns.DISPLAY_NAME},
                null, null, null)) {
            if (c != null && c.moveToFirst()) {
                name = c.getString(0);
            }
        } catch (Exception ignored) {}
        if (name == null) {
            String path = uri.getLastPathSegment();
            name = path != null ? path : "shared_file";
        }
        return name;
    }

    private void evalJs(String js) {
        bridge.getWebView().evaluateJavascript(js, null);
    }

    private void pollAndExecute(String js) {
        pollAndExecuteWithRetry(js, 0);
    }

    private void pollAndExecuteWithRetry(String js, int attempt) {
        if (attempt > 50) return;
        bridge.getWebView().evaluateJavascript(
                "typeof window.__handleShareIntent === 'function'",
                value -> {
                    if ("true".equals(value)) {
                        bridge.getWebView().evaluateJavascript(js, null);
                    } else {
                        mainHandler.postDelayed(
                                () -> pollAndExecuteWithRetry(js, attempt + 1), 100);
                    }
                });
    }

    private static String jsonString(String s) {
        StringBuilder sb = new StringBuilder("\"");
        for (int i = 0; i < s.length(); i++) {
            char c = s.charAt(i);
            switch (c) {
                case '"': sb.append("\\\""); break;
                case '\\': sb.append("\\\\"); break;
                case '\n': sb.append("\\n"); break;
                case '\r': sb.append("\\r"); break;
                case '\t': sb.append("\\t"); break;
                default:
                    if (c < 0x20) {
                        sb.append(String.format("\\u%04x", (int) c));
                    } else {
                        sb.append(c);
                    }
            }
        }
        sb.append("\"");
        return sb.toString();
    }
}
