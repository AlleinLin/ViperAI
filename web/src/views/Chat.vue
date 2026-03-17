<template>
  <div class="chat-page">
    <aside class="sidebar">
      <div class="sidebar-header">
        <h2>对话列表</h2>
        <el-button type="primary" size="small" @click="createNewConversation">
          新对话
        </el-button>
      </div>
      
      <div class="conversation-list">
        <div
          v-for="conv in conversations"
          :key="conv.id"
          :class="['conversation-item', { active: currentConversationId === conv.id }]"
          @click="selectConversation(conv.id)"
        >
          <span>{{ conv.title || `对话 ${conv.id.slice(0, 8)}` }}</span>
        </div>
      </div>
    </aside>
    
    <main class="chat-main">
      <header class="chat-header">
        <el-button @click="$router.push('/home')">返回</el-button>
        <el-button 
          type="success" 
          size="small" 
          :disabled="!currentConversationId || tempSession"
          @click="syncHistory"
        >
          同步历史
        </el-button>
        <div class="model-selector">
          <label>模型：</label>
          <el-select v-model="selectedModel" size="small">
            <el-option label="阿里百炼" value="1" />
            <el-option label="RAG增强" value="2" />
            <el-option label="MCP工具" value="3" />
          </el-select>
        </div>
        <el-checkbox v-model="streamMode">流式输出</el-checkbox>
        <el-button type="primary" size="small" @click="triggerUpload" :loading="uploading">
          上传文档
        </el-button>
        <input
          ref="fileInput"
          type="file"
          accept=".md,.txt"
          style="display: none"
          @change="handleFileUpload"
        />
      </header>
      
      <div class="messages-container" ref="messagesRef">
        <div
          v-for="(msg, index) in messages"
          :key="index"
          :class="['message', msg.isFromUser ? 'user' : 'assistant']"
        >
          <div class="message-avatar">
            {{ msg.isFromUser ? '我' : 'AI' }}
          </div>
          <div class="message-content">
            <div v-html="renderMarkdown(msg.content)"></div>
            <span v-if="msg.streaming" class="streaming-indicator"> ··</span>
            <el-button
              v-if="!msg.isFromUser && !msg.streaming"
              type="text"
              size="small"
              @click="playTTS(msg.content)"
            >
              🔊
            </el-button>
          </div>
        </div>
      </div>
      
      <footer class="chat-footer">
        <el-input
          v-model="inputMessage"
          type="textarea"
          :rows="2"
          placeholder="输入您的问题... (Enter发送, Ctrl+Enter换行)"
          @keydown.enter.exact.prevent="sendMessage"
          :disabled="loading"
        />
        <el-button
          type="primary"
          :loading="loading"
          @click="sendMessage"
        >
          {{ loading ? '发送中...' : '发送' }}
        </el-button>
      </footer>
    </main>
  </div>
</template>

<script>
import { ref, onMounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../utils/api'

export default {
  name: 'ChatView',
  setup() {
    const conversations = ref([])
    const currentConversationId = ref(null)
    const tempSession = ref(false)
    const messages = ref([])
    const inputMessage = ref('')
    const selectedModel = ref('1')
    const streamMode = ref(true)
    const loading = ref(false)
    const uploading = ref(false)
    const messagesRef = ref(null)
    const fileInput = ref(null)
    
    const renderMarkdown = (text) => {
      if (!text && text !== '') return ''
      return String(text)
        .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
        .replace(/\*(.*?)\*/g, '<em>$1</em>')
        .replace(/`(.*?)`/g, '<code class="inline-code">$1</code>')
        .replace(/```(\w*)\n([\s\S]*?)```/g, '<pre><code class="language-$1">$2</code></pre>')
        .replace(/\n/g, '<br>')
    }
    
    const loadConversations = async () => {
      try {
        const response = await api.get('/chat/conversations')
        if (response.data.code === 1000) {
          conversations.value = response.data.data.conversations || []
        }
      } catch (error) {
        console.error('Load conversations error:', error)
      }
    }
    
    const createNewConversation = () => {
      currentConversationId.value = 'temp'
      tempSession.value = true
      messages.value = []
    }
    
    const selectConversation = async (id) => {
      currentConversationId.value = id
      tempSession.value = false
      
      try {
        const response = await api.post('/chat/history', { conversationId: id })
        if (response.data.code === 1000) {
          messages.value = response.data.data.history || []
          scrollToBottom()
        }
      } catch (error) {
        console.error('Load history error:', error)
      }
    }
    
    const syncHistory = async () => {
      if (!currentConversationId.value || tempSession.value) {
        ElMessage.warning('请选择已有会话进行同步')
        return
      }
      
      try {
        const response = await api.post('/chat/history', { conversationId: currentConversationId.value })
        if (response.data.code === 1000) {
          messages.value = response.data.data.history || []
          scrollToBottom()
          ElMessage.success('历史数据同步成功')
        }
      } catch (error) {
        console.error('Sync history error:', error)
        ElMessage.error('同步历史数据失败')
      }
    }
    
    const sendMessage = async () => {
      if (!inputMessage.value.trim() || loading.value) return
      
      const question = inputMessage.value.trim()
      inputMessage.value = ''
      
      messages.value.push({ isFromUser: true, content: question })
      scrollToBottom()
      
      loading.value = true
      
      try {
        if (streamMode.value) {
          await sendStreamMessage(question)
        } else {
          await sendNormalMessage(question)
        }
      } catch (error) {
        console.error('Send message error:', error)
        ElMessage.error('发送失败')
        messages.value.pop()
      } finally {
        loading.value = false
      }
    }
    
    const sendNormalMessage = async (question) => {
      let response
      if (currentConversationId.value === 'temp') {
        response = await api.post('/chat/send-new', {
          question,
          engineType: selectedModel.value
        })
        if (response.data.code === 1000) {
          currentConversationId.value = response.data.data.conversationId
          tempSession.value = false
          loadConversations()
        }
      } else {
        response = await api.post('/chat/send', {
          question,
          engineType: selectedModel.value,
          conversationId: currentConversationId.value
        })
      }
      
      if (response.data.code === 1000) {
        messages.value.push({ isFromUser: false, content: response.data.data.content })
        scrollToBottom()
      } else {
        ElMessage.error(response.data.message || '发送失败')
        messages.value.pop()
      }
    }
    
    const sendStreamMessage = async (question) => {
      const url = currentConversationId.value === 'temp' ? '/chat/stream-new' : '/chat/stream'
      const body = currentConversationId.value === 'temp'
        ? { question, engineType: selectedModel.value }
        : { question, engineType: selectedModel.value, conversationId: currentConversationId.value }
      
      messages.value.push({ isFromUser: false, content: '', streaming: true })
      const aiIndex = messages.value.length - 1
      
      try {
        const response = await fetch('/api/v1' + url, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('token')}`
          },
          body: JSON.stringify(body)
        })
        
        const reader = response.body.getReader()
        const decoder = new TextDecoder()
        let buffer = ''
        
        while (true) {
          const { done, value } = await reader.read()
          if (done) break
          
          buffer += decoder.decode(value, { stream: true })
          const lines = buffer.split('\n')
          buffer = lines.pop() || ''
          
          for (const line of lines) {
            if (line.startsWith('data: ')) {
              const data = line.slice(6).trim()
              if (data === '[DONE]') {
                messages.value[aiIndex].streaming = false
                messages.value = [...messages.value]
                continue
              }
              
              try {
                const parsed = JSON.parse(data)
                if (parsed.conversationId) {
                  currentConversationId.value = parsed.conversationId
                  tempSession.value = false
                  loadConversations()
                }
              } catch {
                messages.value[aiIndex].content += data
                messages.value = [...messages.value]
              }
              
              scrollToBottom()
            }
          }
        }
        
        messages.value[aiIndex].streaming = false
        messages.value = [...messages.value]
      } catch (error) {
        messages.value[aiIndex].streaming = false
        throw error
      }
    }
    
    const playTTS = async (text) => {
      try {
        const createRes = await api.post('/tts/create', { text })
        if (createRes.data.code !== 1000) {
          ElMessage.error('创建语音任务失败')
          return
        }
        
        const taskId = createRes.data.data.taskId
        ElMessage.info('语音合成中，请稍候...')
        
        const maxAttempts = 30
        const pollInterval = 2000
        let attempts = 0
        
        const pollResult = async () => {
          const queryRes = await api.get(`/tts/query?taskId=${taskId}`)
          
          if (queryRes.data.code === 1000) {
            const taskStatus = queryRes.data.data.taskStatus
            
            if (taskStatus === 'Success' && queryRes.data.data.taskResult) {
              const audio = new Audio(queryRes.data.data.taskResult)
              audio.play()
              return true
            } else if (taskStatus === 'Running' || taskStatus === 'Created') {
              attempts++
              if (attempts < maxAttempts) {
                await new Promise(resolve => setTimeout(resolve, pollInterval))
                return await pollResult()
              } else {
                ElMessage.error('语音合成超时')
                return true
              }
            } else {
              ElMessage.error('语音合成失败')
              return true
            }
          }
          
          attempts++
          if (attempts < maxAttempts) {
            await new Promise(resolve => setTimeout(resolve, pollInterval))
            return await pollResult()
          } else {
            ElMessage.error('语音合成超时')
            return true
          }
        }
        
        await pollResult()
      } catch (error) {
        console.error('TTS error:', error)
        ElMessage.error('语音合成请求失败')
      }
    }
    
    const triggerUpload = () => {
      fileInput.value?.click()
    }
    
    const handleFileUpload = async (event) => {
      const file = event.target.files[0]
      if (!file) return
      
      const fileName = file.name.toLowerCase()
      if (!fileName.endsWith('.md') && !fileName.endsWith('.txt')) {
        ElMessage.error('只允许上传 .md 或 .txt 文件')
        event.target.value = ''
        return
      }
      
      const formData = new FormData()
      formData.append('file', file)
      
      uploading.value = true
      try {
        const response = await api.post('/file/upload', formData, {
          headers: { 'Content-Type': 'multipart/form-data' }
        })
        if (response.data.code === 1000) {
          ElMessage.success('文件上传成功，可用于RAG增强模式')
        } else {
          ElMessage.error(response.data.message || '上传失败')
        }
      } catch (error) {
        ElMessage.error('文件上传失败')
      } finally {
        uploading.value = false
        event.target.value = ''
      }
    }
    
    const scrollToBottom = () => {
      nextTick(() => {
        if (messagesRef.value) {
          messagesRef.value.scrollTop = messagesRef.value.scrollHeight
        }
      })
    }
    
    onMounted(() => {
      loadConversations()
    })
    
    return {
      conversations,
      currentConversationId,
      tempSession,
      messages,
      inputMessage,
      selectedModel,
      streamMode,
      loading,
      uploading,
      messagesRef,
      fileInput,
      renderMarkdown,
      createNewConversation,
      selectConversation,
      syncHistory,
      sendMessage,
      playTTS,
      triggerUpload,
      handleFileUpload
    }
  }
}
</script>

<style scoped>
.chat-page {
  height: 100vh;
  display: flex;
  background: #0f0f23;
}

.sidebar {
  width: 280px;
  background: rgba(255, 255, 255, 0.02);
  border-right: 1px solid rgba(255, 255, 255, 0.05);
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.sidebar-header h2 {
  color: #fff;
  font-size: 18px;
  margin: 0;
}

.conversation-list {
  flex: 1;
  overflow-y: auto;
}

.conversation-item {
  padding: 15px 20px;
  color: rgba(255, 255, 255, 0.7);
  cursor: pointer;
  border-bottom: 1px solid rgba(255, 255, 255, 0.02);
  transition: all 0.2s;
}

.conversation-item:hover {
  background: rgba(255, 255, 255, 0.05);
}

.conversation-item.active {
  background: linear-gradient(135deg, rgba(0, 212, 170, 0.2) 0%, rgba(0, 168, 204, 0.2) 100%);
  color: #00d4aa;
}

.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.chat-header {
  padding: 15px 20px;
  background: rgba(255, 255, 255, 0.02);
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  display: flex;
  align-items: center;
  gap: 15px;
}

.model-selector {
  display: flex;
  align-items: center;
  gap: 8px;
  color: rgba(255, 255, 255, 0.7);
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.message {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}

.message.user {
  flex-direction: row-reverse;
}

.message-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 600;
  flex-shrink: 0;
}

.message.user .message-avatar {
  background: linear-gradient(135deg, #00d4aa 0%, #00a8cc 100%);
  color: #fff;
}

.message.assistant .message-avatar {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.message-content {
  max-width: 70%;
  padding: 12px 16px;
  border-radius: 16px;
  color: #fff;
  line-height: 1.6;
}

.message.user .message-content {
  background: linear-gradient(135deg, #00d4aa 0%, #00a8cc 100%);
}

.message.assistant .message-content {
  background: rgba(255, 255, 255, 0.05);
}

.message-content :deep(.inline-code) {
  background: rgba(0, 212, 170, 0.2);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Consolas', monospace;
}

.message-content :deep(pre) {
  background: rgba(0, 0, 0, 0.3);
  padding: 12px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 8px 0;
}

.message-content :deep(code) {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
}

.streaming-indicator {
  color: #00d4aa;
  font-weight: bold;
  animation: blink 1s infinite;
}

@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}

.chat-footer {
  padding: 20px;
  background: rgba(255, 255, 255, 0.02);
  border-top: 1px solid rgba(255, 255, 255, 0.05);
  display: flex;
  gap: 12px;
}

.chat-footer .el-textarea {
  flex: 1;
}

.chat-footer :deep(.el-textarea__inner) {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #fff;
}
</style>
