/* eslint-disable no-undef */
// Dedicated FCM background service worker. It lives beside the vite-plugin-pwa
// workbox SW on its own scope/file, so the two never collide.
//
// Files under public/ are NOT processed by Vite, so env substitution does not
// happen here. The Firebase *web* config is public by design — replace the
// placeholders below with this deployment's values to enable background push.
// Until then the app hides the "Connect" button (see isPushConfigured), so this
// worker is never registered and the placeholder values are never used.
importScripts("https://www.gstatic.com/firebasejs/10.14.1/firebase-app-compat.js")
importScripts("https://www.gstatic.com/firebasejs/10.14.1/firebase-messaging-compat.js")

firebase.initializeApp({
  apiKey: "AIzaSyAkbK44ZDOeg5zdPaIO63rROs5KvkwgFRc",
  projectId: "zoora-a55fa",
  messagingSenderId: "255482847291",
  appId: "1:255482847291:web:3039a7b3405bf481a77c1e",
})

// Background pushes: FCM renders `notification` payloads automatically; this
// handler only needs to exist so the SW claims push scope.
firebase.messaging()
