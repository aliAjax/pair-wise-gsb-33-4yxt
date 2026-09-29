<template>
  <div class="fav-folders">
    <el-row :gutter="16">
      <!-- 左侧：分组列表 -->
      <el-col :xs="24" :sm="8" :md="7" :lg="6">
        <div class="group-list">
          <div class="group-head">
            <span>我的分组</span>
            <el-button size="small" type="primary" text @click="startCreate">
              <el-icon><Plus /></el-icon>&nbsp;新建
            </el-button>
          </div>

          <div
            v-for="g in groupTabs"
            :key="g.id"
            class="group-item"
            :class="{ active: activeGroup === g.id }"
            @click="activeGroup = g.id"
          >
            <el-icon class="folder-icon"><Folder v-if="g.id !== UNCATEGORIZED_GROUP_ID" /><Files v-else /></el-icon>
            <span class="group-name" :title="g.name">{{ g.name }}</span>
            <el-badge :value="g.fav_count" :hidden="g.fav_count === 0" class="group-badge" />
            <el-dropdown
              v-if="g.id !== UNCATEGORIZED_GROUP_ID"
              trigger="click"
              @command="(cmd: string) => onGroupCommand(cmd, g)"
              @click.stop
            >
              <el-icon class="group-more" @click.stop><MoreFilled /></el-icon>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="rename">重命名</el-dropdown-item>
                  <el-dropdown-item command="delete" divided>删除分组</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>

          <div v-if="creating" class="group-create" @click.stop>
            <el-input
              ref="createInputRef"
              v-model="newName"
              size="small"
              placeholder="分组名称"
              maxlength="64"
              @keyup.enter="confirmCreate"
              @blur="cancelCreate"
            />
            <el-button size="small" type="primary" @click="confirmCreate">确定</el-button>
          </div>
        </div>
      </el-col>

      <!-- 右侧：分组下的收藏 -->
      <el-col :xs="24" :sm="16" :md="17" :lg="18">
        <el-table :data="visibleFavorites" empty-text="该分组下暂无收藏" v-loading="loading">
          <el-table-column label="类型" width="92">
            <template #default="{ row }">
              <el-tag size="small" :type="row.target_type === 'plant' ? 'success' : 'warning'">
                {{ FavoriteTargetTypeMap[row.target_type as FavoriteTargetType] }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="标题">
            <template #default="{ row }">
              <router-link
                v-if="targetTitle(row)"
                class="fav-link"
                :to="targetRoute(row)"
              >{{ targetTitle(row) }}</router-link>
              <span v-else class="fav-gone">目标已删除（#{{ row.target_id }}）</span>
            </template>
          </el-table-column>
          <el-table-column label="收藏时间" width="168">
            <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
          </el-table-column>
          <el-table-column label="所属分组" width="170">
            <template #default="{ row }">
              <el-dropdown trigger="click" @command="(gid: number) => onMove(row, gid)">
                <el-button size="small" text>
                  {{ groupNameOf(row.group_id) }}
                  <el-icon class="el-icon--right"><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item
                      v-for="g in moveOptions"
                      :key="g.id"
                      :command="g.id"
                    >{{ g.name }}</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="90">
            <template #default="{ row }">
              <el-button size="small" type="danger" text @click="onRemove(row)">取消收藏</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, Files, Folder, MoreFilled, Plus } from '@element-plus/icons-vue'
import {
  FavoriteTargetTypeMap,
  UNCATEGORIZED_GROUP_ID,
  UNCATEGORIZED_GROUP_NAME,
  type Favorite,
  type FavoriteGroup,
  type FavoriteTargetType,
} from '@/constants/favorite'
import {
  createFavoriteGroup,
  deleteFavoriteGroup,
  listFavoriteGroups,
  listFavorites,
  moveFavorite,
  removeFavorite,
  renameFavoriteGroup,
} from '@/api/favorite'
import { getPlant } from '@/api/plant'
import { getArticle } from '@/api/article'
import { formatDate } from '@/utils/dateFormat'

const loading = ref(false)
const favorites = ref<Favorite[]>([])
const groups = ref<FavoriteGroup[]>([])
const activeGroup = ref<number>(UNCATEGORIZED_GROUP_ID)
const creating = ref(false)
const newName = ref('')
const createInputRef = ref<{ focus: () => void } | null>(null)

// Title/link lookups for favorited targets, fetched once per unique target.
const titles = ref<Record<string, string>>({})

function targetKey(f: Favorite) {
  return `${f.target_type}:${f.target_id}`
}

async function loadAll() {
  loading.value = true
  try {
    const [favs, grps] = await Promise.all([listFavorites(), listFavoriteGroups()])
    favorites.value = favs
    groups.value = grps
    await resolveTitles(favs)
  } finally {
    loading.value = false
  }
}

async function resolveTitles(favs: Favorite[]) {
  const uncached = favs.filter((f) => !(targetKey(f) in titles.value))
  await Promise.all(
    uncached.map(async (f) => {
      const key = targetKey(f)
      try {
        if (f.target_type === 'plant') {
          const p = await getPlant(f.target_id)
          titles.value[key] = p.name
        } else {
          const a = await getArticle(f.target_id)
          titles.value[key] = a.title
        }
      } catch {
        titles.value[key] = ''
      }
    }),
  )
}

const uncategorizedCount = computed(
  () => favorites.value.filter((f) => f.group_id == null).length,
)

const groupTabs = computed(() => {
  const real = groups.value.map((g) => {
    // Trust live favorites as the source of truth for counts.
    const count = favorites.value.filter((f) => f.group_id === g.id).length
    return { id: g.id, name: g.name, fav_count: count }
  })
  return [
    { id: UNCATEGORIZED_GROUP_ID, name: UNCATEGORIZED_GROUP_NAME, fav_count: uncategorizedCount.value },
    ...real,
  ]
})

const moveOptions = computed(() => [
  { id: UNCATEGORIZED_GROUP_ID, name: UNCATEGORIZED_GROUP_NAME },
  ...groups.value.map((g) => ({ id: g.id, name: g.name })),
])

const visibleFavorites = computed(() => {
  if (activeGroup.value === UNCATEGORIZED_GROUP_ID) {
    return favorites.value.filter((f) => f.group_id == null)
  }
  return favorites.value.filter((f) => f.group_id === activeGroup.value)
})

function groupNameOf(groupId: number | null) {
  if (groupId == null) return UNCATEGORIZED_GROUP_NAME
  return groups.value.find((g) => g.id === groupId)?.name ?? UNCATEGORIZED_GROUP_NAME
}

function targetTitle(f: Favorite) {
  return titles.value[targetKey(f)] || ''
}

function targetRoute(f: Favorite) {
  return f.target_type === 'plant' ? `/plants/${f.target_id}` : `/articles/${f.target_id}`
}

async function startCreate() {
  creating.value = true
  newName.value = ''
  await nextTick()
  createInputRef.value?.focus()
}

function cancelCreate() {
  creating.value = false
  newName.value = ''
}

async function confirmCreate() {
  const name = newName.value.trim()
  if (!name) {
    cancelCreate()
    return
  }
  const g = await createFavoriteGroup(name)
  groups.value.push(g)
  creating.value = false
  newName.value = ''
  activeGroup.value = g.id
  ElMessage.success('分组已创建')
}

async function onGroupCommand(cmd: string, g: { id: number; name: string }) {
  if (cmd === 'rename') {
    const { value } = await ElMessageBox.prompt('请输入新的分组名称', '重命名分组', {
      confirmButtonText: '保存',
      cancelButtonText: '取消',
      inputValue: g.name,
      inputValidator: (v) => !!v?.trim() || '分组名称不能为空',
    })
    const updated = await renameFavoriteGroup(g.id, value.trim())
    const idx = groups.value.findIndex((x) => x.id === g.id)
    if (idx >= 0) groups.value[idx] = { ...groups.value[idx], name: updated.name }
    ElMessage.success('分组已重命名')
  } else if (cmd === 'delete') {
    const n = favorites.value.filter((f) => f.group_id === g.id).length
    await ElMessageBox.confirm(
      `确定删除分组「${g.name}」吗？分组内 ${n} 条收藏将移回「未分类」，收藏本身不会被删除。`,
      '删除分组',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
    )
    const res = await deleteFavoriteGroup(g.id)
    groups.value = groups.value.filter((x) => x.id !== g.id)
    // Optimistic state already guaranteed by backend; re-sync group ids.
    favorites.value.forEach((f) => {
      if (f.group_id === g.id) f.group_id = null
    })
    if (activeGroup.value === g.id) activeGroup.value = UNCATEGORIZED_GROUP_ID
    ElMessage.success(`分组已删除，${res.detached ?? n} 条收藏移至未分类`)
  }
}

async function onMove(f: Favorite, groupId: number) {
  const updated = await moveFavorite(f.id, groupId)
  const idx = favorites.value.findIndex((x) => x.id === f.id)
  if (idx >= 0) favorites.value[idx] = { ...favorites.value[idx], group_id: updated.group_id }
  ElMessage.success(groupId === UNCATEGORIZED_GROUP_ID ? '已移至未分类' : '已调整分组')
}

async function onRemove(f: Favorite) {
  await removeFavorite(f.target_type, f.target_id)
  favorites.value = favorites.value.filter((x) => x.id !== f.id)
  ElMessage.success('已取消收藏')
}

onMounted(loadAll)
</script>

<style scoped>
.group-list {
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  padding: 8px;
  background: #fafcf9;
}
.group-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 8px 8px;
  font-weight: 600;
  color: #303133;
}
.group-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: 4px;
  cursor: pointer;
  color: #606266;
}
.group-item:hover { background: #eef5ea; }
.group-item.active { background: #e2efdc; color: #2f7a4d; font-weight: 600; }
.folder-icon { flex: none; }
.group-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.group-badge { margin-right: 4px; }
.group-more { padding: 4px; border-radius: 4px; }
.group-more:hover { background: #d7e7d2; }
.group-create {
  display: flex;
  gap: 6px;
  padding: 6px 8px;
}
.fav-link { color: #2f7a4d; text-decoration: none; }
.fav-link:hover { text-decoration: underline; }
.fav-gone { color: #909399; }
</style>
