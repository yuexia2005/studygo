<template>
  <div class="app">
    <!-- Toast -->
    <div v-if="toast" class="toast" :class="{ error: isError }">{{ toast }}</div>

    <!-- Auth Screen -->
    <div v-if="!token" class="auth-screen">
      <AuthCard
        :loginFn="doLogin"
        :registerFn="doRegister"
      />
    </div>

    <!-- Main Screen -->
    <div v-else class="main-screen">
      <!-- Top Bar -->
      <header class="topbar">
        <div class="tabs">
          <button :class="{ active: mode === 'feed' }" @click="switchMode('feed')">最新</button>
          <button :class="{ active: mode === 'hot' }" @click="switchMode('hot')">热榜</button>
        </div>
        <div class="user-area">
          <span class="username" v-if="username">{{ username }}</span>
          <button class="btn outline small" @click="doLogout">退出</button>
        </div>
      </header>

      <!-- Upload -->
      <VideoUpload ref="uploadRef" :uploadFn="doUpload" />

      <!-- Feed / Hot -->
      <div class="content-section">
        <div class="section-title">{{ mode === 'feed' ? '# 最新视频' : '# 热榜' }}</div>
        <FeedList
          v-if="mode === 'feed'"
          :videos="feedVideos"
          :hasMore="feedHasMore"
          :loading="feedLoading"
          :currentUserId="currentUserId"
          @loadMore="loadFeed"
          @like="(id) => doLikeFeed(id)"
          @delete="(id) => doDeleteFeed(id)"
        />
        <HotList
          v-else
          :videos="hotVideos"
          :loading="hotLoading"
          :currentUserId="currentUserId"
          @like="(id) => doLikeHot(id)"
          @delete="(id) => doDeleteHot(id)"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import AuthCard from './components/AuthCard.vue'
import VideoUpload from './components/VideoUpload.vue'
import FeedList from './components/FeedList.vue'
import HotList from './components/HotList.vue'
import { useAuth, useFeed, useHot, toastMsg, toastError, showToast } from './composables/index.js'
import { getToken } from './api/index.js'

const { token, currentUserId, username, login, register, logout } = useAuth()
const {
  videos: feedVideos, hasMore: feedHasMore, loading: feedLoading,
  loadFeed, resetFeed, toggleLike: toggleLikeFeed, deleteVideo: deleteVideoFeed
} = useFeed()
const {
  videos: hotVideos, loading: hotLoading,
  loadHot, toggleLike: toggleLikeHot
} = useHot()

const mode = ref('feed')
const toast = toastMsg
const isError = toastError
const uploadRef = ref(null)

onMounted(() => {
  if (getToken()) {
    loadFeed()
  }
})

async function doLogin(username, password) {
  await login(username, password)
  showToast('登录成功')
}

async function doRegister(username, password) {
  await register(username, password)
  showToast('注册成功，请登录')
}

function doLogout() {
  logout()
  feedVideos.value = []
  hotVideos.value = []
}

function switchMode(m) {
  mode.value = m
  if (m === 'hot' && hotVideos.value.length === 0) loadHot()
}

async function doUpload(title, description, file) {
  const fd = new FormData()
  fd.append('title', title)
  fd.append('description', description)
  fd.append('video', file)
  const { api } = await import('./api/index.js')
  await api('/api/video/upload', 'POST', fd, true)
  showToast('上传成功')
  if (mode.value === 'feed') { resetFeed(); loadFeed() }
  else { hotVideos.value = []; loadHot() }
}

async function doLikeFeed(id) { try { await toggleLikeFeed(id) } catch (e) { showToast(e.message, true) } }
async function doDeleteFeed(id) { try { await deleteVideoFeed(id); showToast('删除成功') } catch (e) { showToast(e.message, true) } }
async function doLikeHot(id) { try { await toggleLikeHot(id) } catch (e) { showToast(e.message, true) } }
async function doDeleteHot(id) {
  try {
    const { api } = await import('./api/index.js')
    await api(`/api/video/${id}`, 'DELETE')
    hotVideos.value = hotVideos.value.filter(v => v.ID !== id)
    showToast('删除成功')
  } catch (e) { showToast(e.message, true) }
}
</script>

<style>
/* ==========================================
   BLACK / WHITE TECH STYLE — GLOBAL
   ========================================== */

:root {
  --bg: #0a0a0a;
  --surface: #111;
  --surface2: #1a1a1a;
  --border: #2a2a2a;
  --text: #e0e0e0;
  --text-dim: #666;
  --accent: #fff;
  --accent-dim: #888;
  --danger: #333;
  --danger-text: #f66;
  --radius: 2px;
}

* { margin: 0; padding: 0; box-sizing: border-box; }

body {
  background: var(--bg);
  color: var(--text);
  font-family: 'Courier New', 'Source Code Pro', 'Menlo', monospace;
  font-size: 13px;
  line-height: 1.6;
  min-height: 100vh;
}

.app {
  max-width: 680px;
  margin: 0 auto;
  padding: 24px 16px 60px;
}

/* ---- BUTTONS ---- */
.btn {
  display: inline-block;
  padding: 10px 20px;
  border: 1px solid var(--border);
  background: var(--surface2);
  color: var(--text);
  font-family: inherit;
  font-size: 12px;
  letter-spacing: 1px;
  text-transform: uppercase;
  cursor: pointer;
  transition: all .15s;
  width: 100%;
}
.btn:hover { border-color: var(--accent-dim); color: var(--accent); }
.btn.primary { border-color: var(--accent-dim); }
.btn.primary:hover { background: var(--accent); color: var(--bg); }
.btn.outline { background: transparent; }
.btn.danger { border-color: var(--danger); }
.btn.danger:hover { background: var(--danger); color: var(--danger-text); }
.btn.small { width: auto; padding: 6px 14px; font-size: 11px; }

/* ---- INPUTS ---- */
.input-group { margin-bottom: 14px; }
.input-group label {
  display: block;
  color: var(--text-dim);
  font-size: 10px;
  letter-spacing: 2px;
  margin-bottom: 6px;
}
input, textarea {
  width: 100%;
  padding: 10px 12px;
  background: var(--bg);
  border: 1px solid var(--border);
  color: var(--text);
  font-family: inherit;
  font-size: 13px;
  outline: none;
  transition: border .15s;
}
input:focus, textarea:focus { border-color: var(--accent-dim); }
textarea { resize: vertical; min-height: 40px; }
input[type="file"] { display: none; }
.file-label {
  display: block;
  padding: 14px 12px;
  border: 1px dashed var(--border);
  text-align: center;
  cursor: pointer;
  color: var(--text-dim);
  font-size: 11px;
  letter-spacing: 1px;
}
.file-label:hover { border-color: var(--accent-dim); }
.file-label.selected { border-style: solid; color: var(--text); }

/* ---- AUTH ---- */
.auth-screen {
  display: flex; justify-content: center; padding-top: 80px;
}
.auth-card {
  width: 100%; max-width: 400px;
  border: 1px solid var(--border);
  padding: 40px 32px;
  background: var(--surface);
}
.brand { text-align: center; margin-bottom: 24px; font-size: 16px; letter-spacing: 6px; }
.brand .logo { font-size: 28px; display: block; margin-bottom: 8px; }
.divider { height: 1px; background: var(--border); margin-bottom: 24px; }
.auth-form .btn { margin-top: 12px; }
.switch {
  text-align: center; margin-top: 18px; color: var(--text-dim);
  cursor: pointer; font-size: 11px; letter-spacing: 1px;
}
.switch:hover { color: var(--text); }
.error { color: var(--danger-text); text-align: center; margin-top: 14px; font-size: 11px; }

/* ---- TOPBAR ---- */
.topbar {
  display: flex; justify-content: space-between; align-items: center;
  margin-bottom: 24px; padding-bottom: 16px;
  border-bottom: 1px solid var(--border);
}
.tabs { display: flex; gap: 0; }
.tabs button {
  padding: 8px 18px; border: 1px solid var(--border); background: transparent;
  color: var(--text-dim); font-family: inherit; font-size: 11px; letter-spacing: 2px;
  cursor: pointer; margin-right: -1px;
}
.tabs button.active { color: var(--accent); border-color: var(--accent-dim); background: var(--surface2); }
.tabs button:hover { color: var(--text); }

.user-area { display: flex; align-items: center; gap: 12px; }
.username { color: var(--accent); font-size: 12px; font-weight: 600; }

/* ---- SECTION ---- */
.section-title {
  font-size: 11px; letter-spacing: 3px; color: var(--text-dim);
  margin-bottom: 16px; padding-bottom: 8px;
  border-bottom: 1px solid var(--border);
}

/* ---- UPLOAD CARD ---- */
.upload-card {
  border: 1px solid var(--border);
  padding: 20px;
  margin-bottom: 32px;
  background: var(--surface);
}
.upload-card .section-title { margin-top: 0; border: none; padding: 0; margin-bottom: 16px; }

/* ---- CONTENT SECTION ---- */
.content-section {
  border: 1px solid var(--border);
  padding: 20px;
  background: var(--surface);
}

/* ---- VIDEO CARD ---- */
.video-card {
  border: 1px solid var(--border);
  padding: 18px;
  margin-bottom: 16px;
  background: var(--surface2);
  transition: border .15s;
}
.video-card:hover { border-color: var(--accent-dim); }
.vc-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 10px; }
.vc-title { font-size: 15px; font-weight: 700; color: var(--accent); word-break: break-word; }
.vc-desc {
  color: var(--text-dim); font-size: 12px; margin-bottom: 12px;
  padding: 8px 10px; background: var(--bg); border-left: 2px solid var(--accent-dim);
}
.vc-video { width: 100%; border-radius: 0; margin-bottom: 12px; background: #000; }
.vc-meta {
  display: flex; justify-content: space-between;
  font-size: 10px; color: var(--text-dim); letter-spacing: 1px; margin-bottom: 12px;
}
.vc-actions { margin-bottom: 12px; }
.like-btn {
  background: none; border: 1px solid var(--border);
  color: var(--text-dim); font-family: inherit; font-size: 12px;
  padding: 6px 14px; cursor: pointer; letter-spacing: 1px;
}
.like-btn:hover { border-color: var(--accent-dim); }
.like-btn.liked { color: #f44; border-color: #f44; }
.like-icon { margin-right: 4px; }

/* ---- COMMENTS ---- */
.comment-toggle {
  color: var(--text-dim); font-size: 10px; letter-spacing: 2px;
  cursor: pointer; padding: 6px 0; border-top: 1px solid var(--border);
}
.comment-toggle:hover { color: var(--text); }
.comment-section { margin-top: 10px; }
.comment-input-row { display: flex; gap: 8px; margin-bottom: 12px; align-items: flex-start; }
.cmt-input {
  flex: 1; padding: 8px 10px; background: var(--bg);
  border: 1px solid var(--border); color: var(--text);
  font-family: inherit; font-size: 12px; resize: vertical;
}
.cmt-input:focus { border-color: var(--accent-dim); outline: none; }
.comment-list { max-height: 260px; overflow-y: auto; }
.comment-item {
  padding: 10px 0; border-bottom: 1px solid var(--border);
}
.cmt-header { display: flex; justify-content: space-between; margin-bottom: 4px; }
.cmt-author { color: var(--text); font-weight: bold; font-size: 11px; }
.cmt-time { color: var(--text-dim); font-size: 9px; }
.cmt-body { color: var(--text-dim); font-size: 12px; word-break: break-word; }
.cmt-empty { color: var(--text-dim); font-size: 11px; padding: 14px 0; text-align: center; }
.cmt-del {
  margin-top: 4px; background: none; border: 1px solid var(--border);
  color: var(--danger-text); font-family: inherit; font-size: 9px;
  padding: 2px 8px; cursor: pointer; letter-spacing: 1px;
}
.cmt-del:hover { background: var(--danger-text); color: var(--bg); }
.cmt-more {
  text-align: center; color: var(--text-dim); font-size: 10px;
  padding: 10px 0; cursor: pointer; letter-spacing: 2px;
}
.cmt-more:hover { color: var(--text); }

/* ---- STATUS ---- */
.status { text-align: center; color: var(--text-dim); padding: 30px; font-size: 11px; letter-spacing: 2px; }

/* ---- TOAST ---- */
.toast {
  position: fixed; bottom: 30px; left: 50%; transform: translateX(-50%);
  background: var(--surface); color: var(--text);
  border: 1px solid var(--accent-dim);
  padding: 10px 24px; z-index: 999;
  font-size: 11px; letter-spacing: 2px;
}
.toast.error { border-color: var(--danger-text); color: var(--danger-text); }

/* ---- SCROLLBAR ---- */
::-webkit-scrollbar { width: 4px; }
::-webkit-scrollbar-track { background: var(--bg); }
::-webkit-scrollbar-thumb { background: var(--border); }
::-webkit-scrollbar-thumb:hover { background: var(--accent-dim); }
</style>
