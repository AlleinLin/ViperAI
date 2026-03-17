<template>
  <div class="vision-page">
    <header class="vision-header">
      <el-button @click="$router.push('/home')">返回</el-button>
      <h2>图像识别</h2>
    </header>
    
    <main class="vision-main">
      <div class="upload-section">
        <el-upload
          drag
          action="#"
          :auto-upload="false"
          :show-file-list="false"
          accept="image/*"
          @change="handleFileChange"
        >
          <el-icon size="60" color="#00d4aa"><Upload /></el-icon>
          <div class="upload-text">
            <p>拖拽图片到此处或点击上传</p>
            <p class="upload-hint">支持 JPG、PNG、GIF 格式</p>
          </div>
        </el-upload>
      </div>
      
      <div v-if="previewUrl" class="preview-section">
        <img :src="previewUrl" alt="Preview" class="preview-image" />
      </div>
      
      <div v-if="result" class="result-section">
        <h3>识别结果</h3>
        <p class="result-text">{{ result }}</p>
      </div>
    </main>
  </div>
</template>

<script>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../utils/api'

export default {
  name: 'VisionView',
  setup() {
    const previewUrl = ref('')
    const result = ref('')
    
    const handleFileChange = async (file) => {
      if (!file.raw) return
      
      previewUrl.value = URL.createObjectURL(file.raw)
      result.value = ''
      
      const formData = new FormData()
      formData.append('image', file.raw)
      
      try {
        const response = await api.post('/image/recognize', formData, {
          headers: { 'Content-Type': 'multipart/form-data' }
        })
        
        if (response.data.code === 1000) {
          result.value = response.data.data.className
        } else {
          ElMessage.error(response.data.message || '识别失败')
        }
      } catch (error) {
        console.error('Recognition error:', error)
        ElMessage.error('识别失败，请重试')
      }
    }
    
    return {
      previewUrl,
      result,
      handleFileChange
    }
  }
}
</script>

<style scoped>
.vision-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #0f0f23 0%, #1a1a3e 50%, #0f0f23 100%);
}

.vision-header {
  padding: 20px 40px;
  background: rgba(255, 255, 255, 0.02);
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  display: flex;
  align-items: center;
  gap: 20px;
}

.vision-header h2 {
  color: #fff;
  margin: 0;
}

.vision-main {
  padding: 40px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 40px;
}

.upload-section {
  width: 100%;
  max-width: 600px;
}

.upload-section :deep(.el-upload-dragger) {
  background: rgba(255, 255, 255, 0.03);
  border: 2px dashed rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  padding: 60px;
}

.upload-section :deep(.el-upload-dragger:hover) {
  border-color: #00d4aa;
}

.upload-text {
  margin-top: 20px;
  text-align: center;
}

.upload-text p {
  color: #fff;
  font-size: 16px;
  margin: 0;
}

.upload-hint {
  color: rgba(255, 255, 255, 0.5);
  font-size: 14px !important;
  margin-top: 8px !important;
}

.preview-section {
  max-width: 600px;
}

.preview-image {
  max-width: 100%;
  max-height: 400px;
  border-radius: 16px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
}

.result-section {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  padding: 30px;
  max-width: 600px;
  width: 100%;
  text-align: center;
}

.result-section h3 {
  color: rgba(255, 255, 255, 0.7);
  font-size: 14px;
  margin-bottom: 16px;
}

.result-text {
  color: #00d4aa;
  font-size: 24px;
  font-weight: 600;
  margin: 0;
}
</style>
