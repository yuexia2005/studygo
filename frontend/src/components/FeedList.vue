<template>
  <div class="feed-list">
    <VideoCard
      v-for="v in videos"
      :key="v.ID"
      :video="v"
      :currentUserId="currentUserId"
      @like="$emit('like', $event)"
      @delete="$emit('delete', $event)"
    />
    <div v-if="loading" class="status">加载中...</div>
    <div v-else-if="videos.length === 0" class="status">暂无视频</div>
    <button v-if="hasMore && !loading" class="btn outline" @click="$emit('loadMore')">加载更多</button>
  </div>
</template>

<script setup>
import VideoCard from './VideoCard.vue'

defineProps({
  videos: Array,
  hasMore: Boolean,
  loading: Boolean,
  currentUserId: [String, Number],
})

defineEmits(['loadMore', 'like', 'delete'])
</script>
