<template>
  <div class="comment-section">
    <!-- Input -->
    <div class="comment-input-row">
      <textarea v-model="content" rows="2" placeholder="写评论..." class="cmt-input"></textarea>
      <button class="btn small" @click="$emit('post')">发送</button>
    </div>

    <!-- List -->
    <div class="comment-list" ref="listEl">
      <div v-if="comments.length === 0 && !loading" class="cmt-empty">暂无评论</div>
      <div v-for="c in comments" :key="c.ID" class="comment-item">
        <div class="cmt-header">
          <span class="cmt-author">{{ c.username }}</span>
          <span class="cmt-time">{{ formatTime(c.CreatedAt) }}</span>
        </div>
        <div class="cmt-body">{{ c.Content }}</div>
        <button
          v-if="canDelete(c.UserID)"
          class="cmt-del"
          @click="$emit('deleteComment', c.ID)"
        >删除</button>
      </div>
      <div v-if="hasMore" class="cmt-more" @click="$emit('loadMore')">
        {{ loading ? '...' : '加载更多' }}
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  comments: { type: Array, default: () => [] },
  hasMore: Boolean,
  loading: Boolean,
  currentUserId: [String, Number],
  videoAuthorId: [String, Number],
})

const emit = defineEmits(['post', 'deleteComment', 'loadMore'])
const content = ref('')

defineExpose({ content })

function formatTime(t) {
  if (!t) return ''
  return new Date(t).toLocaleString()
}

function canDelete(commentUserId) {
  return props.currentUserId && (commentUserId == props.currentUserId || props.videoAuthorId == props.currentUserId)
}
</script>
