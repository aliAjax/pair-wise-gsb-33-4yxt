<template>
  <div class="page">
    <div class="page-head">
      <h1>我的收藏夹</h1>
      <p class="tip">按分组整理品种与养护文章；没有归属的收藏自动放在「未分类」。</p>
    </div>
    <el-card>
      <FavoriteFolders />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import FavoriteFolders from '@/components/common/FavoriteFolders.vue'
</script>

<style scoped>
.page { max-width: 1100px; margin: 0 auto; }
.page-head { margin-bottom: 12px; }
.page-head h1 { margin: 0 0 4px; }
.tip { color: #909399; font-size: 13px; margin: 0; }
</style>
