<template>
  <div class="page">
    <h1>我的收藏夹</h1>
    <el-row :gutter="16">
      <!-- 分组侧栏 -->
      <el-col :xs="24" :md="7" :lg="6">
        <el-card>
          <template #header>
            <div class="group-header">
              <span>收藏分组</span>
              <el-button type="primary" size="small" @click="openCreate">新建分组</el-button>
            </div>
          </template>
          <el-menu :default-active="activeKey" class="group-menu" @select="onSelect">
            <el-menu-item index="all">
              <el-icon><CollectionTag /></el-icon>
              <span>全部收藏</span>
              <el-tag size="small" class="cnt">{{ groupData?.total_favorite_cnt ?? 0 }}</el-tag>
            </el-menu-item>
            <el-menu-item index="uncategorized">
              <el-icon><Files /></el-icon>
              <span>未分类</span>
              <el-tag size="small" type="info" class="cnt">{{ groupData?.uncategorized_cnt ?? 0 }}</el-tag>
            </el-menu-item>
            <el-menu-item v-for="g in groupData?.groups ?? []" :key="g.id" :index="`g-${g.id}`">
              <el-icon><Folder /></el-icon>
              <span class="group-name" :title="g.name">{{ g.name }}</span>
              <el-tag size="small" type="success" class="cnt">{{ g.favorite_cnt }}</el-tag>
              <el-icon class="del-icon" title="删除分组" @click.stop="onDeleteGroup(g)"><Delete /></el-icon>
            </el-menu-item>
          </el-menu>
        </el-card>
      </el-col>

      <!-- 收藏列表 -->
      <el-col :xs="24" :md="17" :lg="18">
        <el-card>
          <template #header>
            <div class="list-header">
              <span>{{ activeTitle }}（{{ favorites.length }}）</span>
              <el-radio-group v-model="typeFilter" size="small" @change="loadFavorites">
                <el-radio-button value="">全部类型</el-radio-button>
                <el-radio-button value="plant">植物品种</el-radio-button>
                <el-radio-button value="article">养护文章</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          <el-table :data="favorites" empty-text="该分组下暂无收藏" v-loading="loading">
            <el-table-column label="收藏内容" min-width="220">
              <template #default="{ row }">
                <el-link type="primary" :underline="false" @click="openTarget(row)">
                  {{ targetLabel(row) }}
                </el-link>
              </template>
            </el-table-column>
            <el-table-column label="类型" width="100">
              <template #default="{ row }">{{ FavoriteTargetTypeMap[row.target_type as FavoriteTargetType] }}</template>
            </el-table-column>
            <el-table-column label="所属分组" width="120">
              <template #default="{ row }">{{ row.group_id ? groupName(row.group_id) : UNCATEGORIZED_GROUP }}</template>
            </el-table-column>
            <el-table-column label="收藏时间" width="110">
              <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="190">
              <template #default="{ row }">
                <el-dropdown trigger="click" @command="chooseGroup(row, $event)">
                  <el-button size="small">移到分组<el-icon class="el-icon--right"><ArrowDown /></el-icon></el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item :command="null">未分类</el-dropdown-item>
                      <el-dropdown-item v-for="g in groupData?.groups ?? []" :key="g.id" :command="g.id">
                        {{ g.name }}
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
                <el-button size="small" type="danger" plain @click="onRemove(row)">取消收藏</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <!-- 新建分组对话框 -->
    <el-dialog v-model="createVisible" title="新建收藏分组" width="360px">
      <el-form @submit.prevent="confirmCreate">
        <el-form-item>
          <el-input v-model="newName" maxlength="64" show-word-limit placeholder="请输入分组名称（同账号下不可重名）" @keyup.enter="confirmCreate" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="confirmCreate">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, CollectionTag, Delete, Files, Folder } from '@element-plus/icons-vue'
import {
  createFavoriteGroup,
  deleteFavoriteGroup,
  listFavoriteGroups,
  listFavorites,
  moveFavorite,
  removeFavorite,
  type FavoriteListParams,
} from '@/api/favorite'
import { getPlant, listPlants } from '@/api/plant'
import { getArticle, listArticles } from '@/api/article'
import {
  FavoriteTargetTypeMap,
  UNCATEGORIZED_GROUP,
  type Favorite,
  type FavoriteGroupListData,
  type FavoriteTargetType,
} from '@/constants/favorite'
import { formatDate } from '@/utils/dateFormat'

const router = useRouter()

const groupData = ref<FavoriteGroupListData | null>(null)
const favorites = ref<Favorite[]>([])
const loading = ref(false)
const typeFilter = ref<'' | FavoriteTargetType>('')

const activeKey = ref<'all' | 'uncategorized' | string>('all')
const activeGroupId = computed<number | null>(() => {
  return typeof activeKey.value === 'string' && activeKey.value.startsWith('g-')
    ? Number(activeKey.value.slice(2))
    : null
})
const activeTitle = computed(() => {
  if (activeKey.value === 'all') return '全部收藏'
  if (activeKey.value === 'uncategorized') return UNCATEGORIZED_GROUP
  return groupName(activeGroupId.value as number) || '收藏'
})

// 目标标题缓存（按品种/文章 id），保证链接可辨识、可打开
const plantNames = ref<Record<number, string>>({})
const articleNames = ref<Record<number, string>>({})

const createVisible = ref(false)
const newName = ref('')
const creating = ref(false)

onMounted(async () => {
  await loadGroups()
  await loadFavorites()
})

async function loadGroups() {
  groupData.value = await listFavoriteGroups()
}

async function loadFavorites() {
  loading.value = true
  try {
    const params: FavoriteListParams = { target_type: typeFilter.value }
    if (activeKey.value === 'uncategorized') params.uncategorized = true
    else if (activeGroupId.value) params.group_id = activeGroupId.value
    favorites.value = await listFavorites(params)
    await resolveTargetNames(favorites.value)
  } finally {
    loading.value = false
  }
}

async function resolveTargetNames(items: Favorite[]) {
  const plantIds = [...new Set(items.filter((i) => i.target_type === 'plant').map((i) => i.target_id))]
  const articleIds = [...new Set(items.filter((i) => i.target_type === 'article').map((i) => i.target_id))]
  const missingPlants = plantIds.filter((id) => !(id in plantNames.value))
  const missingArticles = articleIds.filter((id) => !(id in articleNames.value))

  // 先翻列表接口批量补名；列表未覆盖（如已下架或超过分页）的 id 再按详情补齐
  if (missingPlants.length) {
    let pageNo = 1
    for (;;) {
      const page = await listPlants({ page: pageNo, page_size: 100 })
      page.list.forEach((p) => { plantNames.value[p.id] = p.name })
      if (plantIds.every((id) => id in plantNames.value) || pageNo * 100 >= page.total) break
      pageNo += 1
    }
  }
  if (missingArticles.length) {
    let pageNo = 1
    for (;;) {
      const page = await listArticles({ page: pageNo, page_size: 100 })
      page.list.forEach((a) => { articleNames.value[a.id] = a.title })
      if (articleIds.every((id) => id in articleNames.value) || pageNo * 100 >= page.total) break
      pageNo += 1
    }
  }
  await Promise.all(
    plantIds
      .filter((id) => !(id in plantNames.value))
      .map((id) => getPlant(id).then((p) => { plantNames.value[id] = p.name }).catch(() => undefined)),
  )
  await Promise.all(
    articleIds
      .filter((id) => !(id in articleNames.value))
      .map((id) => getArticle(id).then((a) => { articleNames.value[id] = a.title }).catch(() => undefined)),
  )
}

function groupName(id: number): string {
  return groupData.value?.groups.find((g) => g.id === id)?.name || UNCATEGORIZED_GROUP
}

function targetLabel(row: Favorite): string {
  const map = row.target_type === 'plant' ? plantNames.value : articleNames.value
  const name = map[row.target_id]
  const prefix = FavoriteTargetTypeMap[row.target_type]
  return name ? `【${prefix}】${name}` : `【${prefix}】目标 #${row.target_id}（可能已删除）`
}

function openTarget(row: Favorite) {
  if (row.target_type === 'plant') {
    router.push(`/plants/${row.target_id}`)
  } else {
    router.push(`/articles/${row.target_id}`)
  }
}

function onSelect(key: string) {
  activeKey.value = key
  loadFavorites()
}

function openCreate() {
  newName.value = ''
  createVisible.value = true
}

async function confirmCreate() {
  const name = newName.value.trim()
  if (!name) {
    ElMessage.warning('请输入分组名称')
    return
  }
  if (name === UNCATEGORIZED_GROUP) {
    ElMessage.warning(`「${UNCATEGORIZED_GROUP}」是系统保留名称`)
    return
  }
  creating.value = true
  try {
    const g = await createFavoriteGroup(name)
    await loadGroups()
    createVisible.value = false
    ElMessage.success(`分组「${g.name}」已创建`)
  } finally {
    creating.value = false
  }
}

async function onDeleteGroup(g: { id: number; name: string; favorite_cnt: number }) {
  try {
    await ElMessageBox.confirm(
      `删除分组「${g.name}」后，其中 ${g.favorite_cnt} 条收藏将回到「${UNCATEGORIZED_GROUP}」，收藏本身不会被删除。`,
      '删除分组',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
    )
  } catch {
    return
  }
  await deleteFavoriteGroup(g.id)
  if (activeKey.value === `g-${g.id}`) activeKey.value = 'uncategorized'
  await loadGroups()
  await loadFavorites()
  ElMessage.success('分组已删除，收藏已移回未分类')
}

function chooseGroup(row: Favorite, command: string | number | null) {
  onMove(row, command === null ? null : Number(command))
}

async function onMove(row: Favorite, groupId: number | null) {
  if ((row.group_id ?? null) === groupId) return
  await moveFavorite(row.target_type, row.target_id, groupId)
  await loadGroups()
  await loadFavorites()
  ElMessage.success(groupId ? `已移至「${groupName(groupId)}」` : `已移至「${UNCATEGORIZED_GROUP}」`)
}

async function onRemove(row: Favorite) {
  await removeFavorite(row.target_type, row.target_id)
  await loadGroups()
  await loadFavorites()
  ElMessage.success('已取消收藏')
}
</script>

<style scoped>
.page { max-width: 1200px; margin: 0 auto; }
.group-header, .list-header { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.group-menu { border-right: none; }
.group-menu .el-menu-item { display: flex; align-items: center; gap: 6px; }
.group-menu .group-name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.cnt { margin-left: auto; }
.del-icon { margin-left: 8px; color: #f56c6c; cursor: pointer; }
.del-icon:hover { opacity: .8; }
</style>
