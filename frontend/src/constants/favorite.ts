export type FavoriteTargetType = 'plant' | 'article'

export const FavoriteTargetTypeMap: Record<FavoriteTargetType, string> = {
  plant: '植物品种',
  article: '养护文章',
}

export interface Favorite {
  id: number
  user_id: number
  target_type: FavoriteTargetType
  target_id: number
  group_id: number | null
  created_at: string
}

export interface FavoriteGroup {
  id: number
  name: string
  fav_count: number
  created_at: string
}

// Pseudo group id used by the UI/API for favorites without a group.
export const UNCATEGORIZED_GROUP_ID = 0
export const UNCATEGORIZED_GROUP_NAME = '未分类'
