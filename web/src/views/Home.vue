<template>
  <div class="home-page">
    <header class="home-header">
      <div class="logo">
        <h1>ViperAI</h1>
      </div>
      <div class="user-actions">
        <el-button type="danger" @click="handleLogout">退出登录</el-button>
      </div>
    </header>
    
    <main class="home-main">
      <div class="welcome-section">
        <h2>欢迎使用 ViperAI</h2>
        <p>探索人工智能的无限可能</p>
      </div>
      
      <div class="feature-grid">
        <div class="feature-card" @click="$router.push('/chat')">
          <div class="feature-icon">
            <el-icon size="48"><ChatDotRound /></el-icon>
          </div>
          <h3>智能对话</h3>
          <p>与AI进行自然语言交互，支持多种模型</p>
        </div>
        
        <div class="feature-card" @click="$router.push('/vision')">
          <div class="feature-icon">
            <el-icon size="48"><Camera /></el-icon>
          </div>
          <h3>图像识别</h3>
          <p>上传图片进行智能识别和分析</p>
        </div>
      </div>
    </main>
  </div>
</template>

<script>
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  name: 'HomeView',
  setup() {
    const router = useRouter()
    
    const handleLogout = async () => {
      try {
        await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        localStorage.removeItem('token')
        ElMessage.success('退出登录成功')
        router.push('/login')
      } catch {
        // User cancelled
      }
    }
    
    return {
      handleLogout
    }
  }
}
</script>

<style scoped>
.home-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #0f0f23 0%, #1a1a3e 50%, #0f0f23 100%);
  position: relative;
  overflow: hidden;
}

.home-page::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: 
    radial-gradient(circle at 20% 30%, rgba(0, 212, 170, 0.08) 0%, transparent 50%),
    radial-gradient(circle at 80% 70%, rgba(0, 168, 204, 0.08) 0%, transparent 50%);
  pointer-events: none;
}

.home-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 40px;
  background: rgba(255, 255, 255, 0.02);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  position: relative;
  z-index: 10;
}

.logo h1 {
  font-size: 28px;
  font-weight: 700;
  background: linear-gradient(135deg, #00d4aa 0%, #00a8cc 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin: 0;
}

.user-actions .el-button {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: #fff;
  border-radius: 10px;
}

.user-actions .el-button:hover {
  background: rgba(255, 100, 100, 0.2);
  border-color: rgba(255, 100, 100, 0.3);
}

.home-main {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: calc(100vh - 80px);
  padding: 40px;
  position: relative;
  z-index: 1;
}

.welcome-section {
  text-align: center;
  margin-bottom: 60px;
}

.welcome-section h2 {
  font-size: 42px;
  font-weight: 700;
  color: #fff;
  margin-bottom: 16px;
}

.welcome-section p {
  font-size: 18px;
  color: rgba(255, 255, 255, 0.6);
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 30px;
  max-width: 900px;
  width: 100%;
}

.feature-card {
  padding: 40px 30px;
  background: rgba(255, 255, 255, 0.03);
  backdrop-filter: blur(20px);
  border-radius: 20px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  cursor: pointer;
  transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  text-align: center;
}

.feature-card:hover {
  transform: translateY(-10px) scale(1.02);
  background: rgba(255, 255, 255, 0.05);
  border-color: rgba(0, 212, 170, 0.3);
  box-shadow: 0 20px 60px rgba(0, 212, 170, 0.15);
}

.feature-icon {
  margin-bottom: 20px;
  color: #00d4aa;
}

.feature-card h3 {
  font-size: 22px;
  font-weight: 600;
  color: #fff;
  margin-bottom: 12px;
}

.feature-card p {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.6);
  line-height: 1.6;
}
</style>
