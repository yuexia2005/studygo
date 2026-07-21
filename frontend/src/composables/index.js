import { ref, reactive } from 'vue'
import { api, getToken, setToken, clearToken, setUserId, getUserId, setUsername, getUsername } from '../api/index.js'

// ============ Auth ============
export function useAuth() {
  const token = ref(getToken())
  const currentUserId = ref(getUserId())
  const username = ref(getUsername())

  async function login(usernameVal, password) {
    const data = await api('/login', 'POST', { username: usernameVal, password })
    setToken(data.token)
    setUserId(data.user_id)
    setUsername(data.username)
    token.value = data.token
    currentUserId.value = data.user_id
    username.value = data.username
  }

  async function register(usernameVal, password) {
    await api('/register', 'POST', { username: usernameVal, password })
  }

  function logout() {
    clearToken()
    token.value = ''
    currentUserId.value = null
    username.value = ''
  }

  return { token, currentUserId, username, login, register, logout }
}

// ============ Toast ============
export const toastMsg = ref('')
export const toastError = ref(false)
let toastTimer = null

export function showToast(msg, isError = false) {
  toastMsg.value = msg
  toastError.value = isError
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toastMsg.value = '' }, 2500)
}

// ============ Feed ============
export function useFeed() {
  const videos = ref([])
  const hasMore = ref(true)
  const loading = ref(false)
  let lastId = 0

  async function loadFeed() {
    if (!hasMore.value || loading.value) return
    loading.value = true
    try {
      let url = `/api/feed?limit=5`
      if (lastId > 0) url += `&last_id=${lastId}`
      const data = await api(url)
      const list = data.list || data
      if (!list || list.length === 0) {
        hasMore.value = false
        return
      }
      videos.value.push(...list)
      lastId = list[list.length - 1].ID
      hasMore.value = list.length === 5
    } finally {
      loading.value = false
    }
  }

  function resetFeed() {
    videos.value = []
    lastId = 0
    hasMore.value = true
  }

  async function toggleLike(videoId) {
    const data = await api(`/api/video/${videoId}/like`, 'POST')
    const video = videos.value.find(v => v.ID === videoId)
    if (video) {
      video.is_liked = data.liked
      video.LikeCount = data.like_count
    }
    return data
  }

  async function deleteVideo(videoId) {
    await api(`/api/video/${videoId}`, 'DELETE')
    videos.value = videos.value.filter(v => v.ID !== videoId)
  }

  return { videos, hasMore, loading, loadFeed, resetFeed, toggleLike, deleteVideo }
}

// ============ Hot ============
export function useHot() {
  const videos = ref([])
  const loading = ref(false)

  async function loadHot() {
    loading.value = true
    try {
      const data = await api('/api/hot?limit=10')
      videos.value = data.list || []
    } finally {
      loading.value = false
    }
  }

  async function toggleLike(videoId) {
    const data = await api(`/api/video/${videoId}/like`, 'POST')
    const video = videos.value.find(v => v.ID === videoId)
    if (video) {
      video.is_liked = data.liked
      video.LikeCount = data.like_count
    }
    return data
  }

  return { videos, loading, loadHot, toggleLike }
}

// ============ Upload ============
export function useUpload() {
  const title = ref('')
  const description = ref('')
  const file = ref(null)

  function resetForm() {
    title.value = ''
    description.value = ''
    file.value = null
  }

  async function upload() {
    if (!title.value || !file.value) throw new Error('标题和视频文件不能为空')
    const fd = new FormData()
    fd.append('title', title.value)
    fd.append('description', description.value)
    fd.append('video', file.value)
    await api('/api/video/upload', 'POST', fd, true)
  }

  return { title, description, file, resetForm, upload }
}

// ============ Comments ============
export function useComments(videoId) {
  const comments = ref([])
  const hasMore = ref(true)
  const loading = ref(false)
  const content = ref('')
  let lastId = 0

  async function loadComments(isLoadMore = false) {
    if (loading.value) return
    loading.value = true
    try {
      let url = `/api/video/${videoId}/comments?limit=5`
      if (isLoadMore && lastId > 0) url += `&last_id=${lastId}`
      const data = await api(url)
      if (!isLoadMore) {
        comments.value = data.comments || []
      } else {
        comments.value.push(...(data.comments || []))
      }
      hasMore.value = data.has_more
      lastId = data.last_id
    } finally {
      loading.value = false
    }
  }

  async function postComment() {
    if (!content.value.trim()) return
    await api(`/api/video/${videoId}/comment`, 'POST', { content: content.value })
    content.value = ''
    lastId = 0
    hasMore.value = true
    await loadComments()
  }

  async function deleteComment(commentId) {
    await api(`/api/comment/${commentId}`, 'DELETE')
    lastId = 0
    hasMore.value = true
    await loadComments()
  }

  return { comments, hasMore, loading, content, loadComments, postComment, deleteComment }
}
