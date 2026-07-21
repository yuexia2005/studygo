<template>
  <div class="video-card">
    <div class="vc-header">
      <span class="vc-title">{{ video.Title }}</span>
      <button v-if="isAuthor" class="btn danger small" @click="$emit('delete', video.ID)">删除</button>
    </div>
    <p v-if="video.Description" class="vc-desc">{{ video.Description }}</p>
    <video controls class="vc-video" :src="videoUrl"></video>
    <div class="vc-meta">
      <span>用户: {{ video.username || '#' + video.UserID }}</span>
      <span>{{ formatTime(video.CreatedAt) }}</span>
    </div>
    <div class="vc-actions">
      <button class="like-btn" :class="{ liked: video.is_liked }" @click="$emit('like', video.ID)">
        <span class="like-icon">{{ video.is_liked ? '&#9829;' : '&#9825;' }}</span>
        <span>{{ video.LikeCount || 0 }}</span>
      </button>
    </div>
    <div class="comment-toggle" @click="showComments = !showComments">
      评论 {{ showComments ? '&#9650;' : '&#9660;' }}
    </div>
    <CommentSection
      v-if="showComments"
      :comments="comments"
      :hasMore="commentHasMore"
      :loading="commentLoading"
      :currentUserId="currentUserId"
      :videoAuthorId="video.UserID"
      @post="handlePost"
      @deleteComment="handleDeleteComment"
      @loadMore="handleLoadMore"
      ref="commentRef"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import CommentSection from './CommentSection.vue'
import { useComments } from '../composables/index.js'

const props = defineProps({
  video: Object,
  currentUserId: [String, Number],
})

const emit = defineEmits(['like', 'delete'])

const showComments = ref(false)
const commentRef = ref(null)
const { comments, hasMore, loading, content, loadComments, postComment, deleteComment } = useComments(props.video.ID)

const commentHasMore = computed(() => hasMore.value)
const commentLoading = computed(() => loading.value)

const isAuthor = computed(() => props.currentUserId && props.video.UserID == props.currentUserId)

const videoUrl = computed(() => {
  const name = props.video.FilePath?.split('/').pop() || ''
  return `/uploads/${name}`
})

watch(showComments, (val) => {
  if (val && comments.value.length === 0) loadComments()
})

async function handlePost() {
  const refEl = commentRef.value
  if (!refEl?.content) return
  content.value = refEl.content
  await postComment()
}
async function handleDeleteComment(id) { await deleteComment(id) }
function handleLoadMore() { loadComments(true) }

function formatTime(t) {
  if (!t) return ''
  return new Date(t).toLocaleString()
}
</script>
