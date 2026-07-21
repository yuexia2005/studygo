<template>
  <div class="upload-card">
    <div class="section-title">+ 上传视频</div>
    <div class="input-group">
      <label>标题</label>
      <input v-model="title" placeholder="请输入视频标题" />
    </div>
    <div class="input-group">
      <label>描述</label>
      <textarea v-model="description" rows="2" placeholder="请输入视频描述"></textarea>
    </div>
    <div class="input-group">
      <label>文件</label>
      <label class="file-label" :class="{ selected: fileName }" for="videoFile">
        {{ fileName || '选择 MP4 文件' }}
      </label>
      <input id="videoFile" type="file" accept="video/mp4" @change="onFileChange" ref="fileInput" />
    </div>
    <button class="btn primary" @click="handleUpload" :disabled="uploading">
      {{ uploading ? '上传中...' : '上传' }}
    </button>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { showToast } from '../composables/index.js'

const props = defineProps({
  uploadFn: Function,
})

const title = ref('')
const description = ref('')
const file = ref(null)
const fileName = ref('')
const uploading = ref(false)
const fileInput = ref(null)

function onFileChange(e) {
  file.value = e.target.files[0]
  fileName.value = file.value?.name || ''
}

async function handleUpload() {
  if (!title.value || !file.value) {
    showToast('标题和视频文件不能为空', true)
    return
  }
  uploading.value = true
  try {
    await props.uploadFn(title.value, description.value, file.value)
    clear()
  } catch (e) {
    showToast(e.message, true)
  } finally {
    uploading.value = false
  }
}

function clear() {
  title.value = ''
  description.value = ''
  file.value = null
  fileName.value = ''
  if (fileInput.value) fileInput.value.value = ''
}

defineExpose({ clear })
</script>
