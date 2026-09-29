<template>
  <div class="page">
    <el-row :gutter="16">
      <el-col :xs="24" :md="10">
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
      <el-col :xs="24" :md="14">
        <el-card>
          <template #header>
            <div class="fav-header">
              <span>我的收藏</span>
              <el-button type="primary" link @click="$router.push('/favorites')">
                管理分组<el-icon class="el-icon--right"><ArrowRight /></el-icon>
              </el-button>
            </div>
          </template>
          <el-table :data="favorites" empty-text="暂无收藏">
            <el-table-column label="类型" width="100">
              <template #default="{ row }">{{ FavoriteTargetTypeMap[row.target_type as FavoriteTargetType] }}</template>
            </el-table-column>
            <el-table-column prop="target_id" label="目标 ID" />
            <el-table-column label="所属分组">
              <template #default="{ row }">{{ row.group_id ? groupName(row.group_id) : UNCATEGORIZED_GROUP }}</template>
            </el-table-column>
            <el-table-column label="时间">
              <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
            </el-table-column>
          </el-table>
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
import { useUserStore } from '@/stores/userStore'
import { listFavorites, listFavoriteGroups } from '@/api/favorite'
import { FavoriteTargetTypeMap, UNCATEGORIZED_GROUP, type Favorite, type FavoriteGroupListData, type FavoriteTargetType } from '@/constants/favorite'
import { formatDate } from '@/utils/dateFormat'
import type { UserInfo } from '@/types/api'

const userStore = useUserStore()
const profile = ref<UserInfo | null>(null)
const favorites = ref<Favorite[]>([])
const groupData = ref<FavoriteGroupListData | null>(null)
const form = reactive({ nickname: '', bio: '', avatar: '' })

onMounted(async () => {
  await userStore.load()
  profile.value = userStore.profile
  form.nickname = profile.value?.nickname || ''
  form.bio = profile.value?.bio || ''
  form.avatar = profile.value?.avatar || ''
  const [favList, groups] = await Promise.all([listFavorites(), listFavoriteGroups()])
  favorites.value = favList
  groupData.value = groups
})

function groupName(id: number): string {
  return groupData.value?.groups.find((g) => g.id === id)?.name || UNCATEGORIZED_GROUP
}

async function save() {
  await userStore.save({ nickname: form.nickname, bio: form.bio, avatar: form.avatar })
  profile.value = userStore.profile
  ElMessage.success('资料已更新')
}
</script>

<style scoped>
.page { max-width: 1000px; margin: 0 auto; }
.avatar { margin-bottom: 12px; }
.profile-form { margin-top: 12px; }
.fav-header { display: flex; align-items: center; justify-content: space-between; }
</style>
