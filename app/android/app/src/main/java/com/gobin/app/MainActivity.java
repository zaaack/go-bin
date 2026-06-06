package com.gobin.app;

import android.content.Intent;
import android.net.Uri;
import android.os.Bundle;
import android.util.Base64;
import android.webkit.MimeTypeMap;

import com.getcapacitor.BridgeActivity;

import java.io.ByteArrayOutputStream;
import java.io.InputStream;
import java.util.ArrayList;

public class MainActivity extends BridgeActivity {

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

        // Wait for bridge to be ready, then evaluate JS
        bridge.getActivity().runOnUiThread(() -> {
            try {
                Thread.sleep(500);
            } catch (InterruptedException ignored) {}

            if (Intent.ACTION_SEND.equals(action)) {
                handleSendSingle(intent, type);
            } else {
                handleSendMultiple(intent);
            }
        });
    }

    private void handleSendSingle(Intent intent, String type) {
        String sharedText = intent.getStringExtra(Intent.EXTRA_TEXT);

        Uri fileUri = intent.getParcelableExtra(Intent.EXTRA_STREAM);
        if (fileUri != null) {
            String fileJson = uriToFileJson(fileUri);
            if (fileJson != null) {
                String escapedText = sharedText != null ? escapeJs(sharedText) : "undefined";
                String js = "if(window.__handleShareIntent){window.__handleShareIntent("
                        + escapedText + ",[" + fileJson + "]);}";
                bridge.getWebView().evaluateJavascript(js, null);
                return;
            }
        }

        if (sharedText != null) {
            String escaped = escapeJs(sharedText);
            String js = "if(window.__handleShareIntent){window.__handleShareIntent("
                    + escaped + ");}";
            bridge.getWebView().evaluateJavascript(js, null);
        }
    }

    private void handleSendMultiple(Intent intent) {
        ArrayList<Uri> uris = intent.getParcelableArrayListExtra(Intent.EXTRA_STREAM);
        if (uris == null || uris.isEmpty()) return;

        StringBuilder sb = new StringBuilder("[");
        for (int i = 0; i < uris.size(); i++) {
            if (i > 0) sb.append(",");
            String json = uriToFileJson(uris.get(i));
            if (json != null) sb.append(json);
        }
        sb.append("]");

        String js = "if(window.__handleShareIntent){window.__handleShareIntent(undefined,"
                + sb.toString() + ");}";
        bridge.getWebView().evaluateJavascript(js, null);
    }

    private String uriToFileJson(Uri uri) {
        try {
            String name = getFileName(uri);
            String mimeType = getContentResolver().getType(uri);
            if (mimeType == null) {
                mimeType = MimeTypeMap.getSingleton().getMimeTypeFromExtension(
                        name.substring(name.lastIndexOf('.') + 1));
            }
            if (mimeType == null) mimeType = "application/octet-stream";

            InputStream is = getContentResolver().openInputStream(uri);
            if (is == null) return null;

            ByteArrayOutputStream baos = new ByteArrayOutputStream();
            byte[] buf = new byte[8192];
            int n;
            while ((n = is.read(buf)) != -1) {
                baos.write(buf, 0, n);
            }
            is.close();

            String base64 = Base64.encodeToString(baos.toByteArray(), Base64.NO_WRAP);
            String dataUrl = "data:" + mimeType + ";base64," + base64;

            return "{\"name\":" + jsonString(name)
                    + ",\"uri\":" + jsonString(uri.toString())
                    + ",\"dataUrl\":" + jsonString(dataUrl) + "}";
        } catch (Exception e) {
            return null;
        }
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

    private static String escapeJs(String s) {
        return jsonString(s);
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
