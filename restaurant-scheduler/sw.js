// 空的 Service Worker 占位文件
// 避免浏览器缓存或扩展导致的 sw.js 404 错误

self.addEventListener('install', (event) => {
    self.skipWaiting();
});

self.addEventListener('activate', (event) => {
    event.waitUntil(clients.claim());
});

self.addEventListener('fetch', (event) => {
    // 不缓存任何请求，直接网络获取
    event.respondWith(fetch(event.request));
});
