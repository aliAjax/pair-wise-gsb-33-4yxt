<template>
  <div class="page">
    <el-row :gutter="16">
      <el-col :xs="24" :md="9">
        <el-card>
          <template #header>个人资料</template>
          <el-avatar :size="80" :src="profile?.avatar" class="avatar">{{ profile?.nickname?.[0] }}</el-avatar>
          <el-form label-width="80px" class="profile-form">
            <el-form-item label="用户名"><el-input :model-value="profile?.username" disabled /></el-form-item>
            <el-form-item label="昵称"><el-input v-model="form.nickname" /></el-form-item>
            <el-form-item label="邮箱"><el-input :model-value="profile?.email" disabled /></el-form-item>
            <el-form-item label="简介"><el-input v-model="form.bio" type="textarea" :rows="3" /></el-form-item>
            <el-form-item label="头像"><ImageUploader v-model="form.avatar" /></el-form-item>
            <el-form-item>
              <el-button type="primary" @click="save">保存修改</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="15">
        <el-card>
          <template #header>
            <div class="fav-header">
              <span>我的收藏</span>
              <el-button type="primary" text @click="$router.push('/favorites')">
                管理收藏夹<el-icon class="el-icon--right"><ArrowRight /></el-icon>
              </el-button>
            </div>
          </template>
          <FavoriteFolders />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { ArrowRight } from '@element-plus/icons-vue'
import ImageUploader from '@/components/common/ImageUploader.vue'
import FavoriteFolders from '@/components/common/FavoriteFolders.vue'
import { useUserStore } from '@/stores/userStore'
import type { UserInfo } from '@/types/api'

const userStore = useUserStore()
const profile = ref<UserInfo | null>(null)
const form = reactive({ nickname: '', bio: '', avatar: '' })

onMounted(async () => {
  await userStore.load()
  profile.value = userStore.profile
  form.nickname = profile.value?.nickname || ''
  form.bio = profile.value?.bio || ''
  form.avatar = profile.value?.avatar || ''
})

async function save() {
  await userStore.save({ nickname: form.nickname, bio: form.bio, avatar: form.avatar })
  profile.value = userStore.profile
  ElMessage.success('资料已更新')
}
</script>

<style scoped>
.page { max-width: 1100px; margin: 0 auto; }
.avatar { margin-bottom: 12px; }
.profile-form { margin-top: 12px; }
.fav-header { display: flex; justify-content: space-between; align-items: center; }
</style>
