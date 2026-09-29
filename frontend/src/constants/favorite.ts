export type FavoriteTargetType = 'plant' | 'article'

export const FavoriteTargetTypeMap: Record<FavoriteTargetType, string> = {
  plant: '植物品种',
  article: '养护文章',
}

// Virtual bucket shown for favorites without a custom group.
export const UNCATEGORIZED_GROUP = '未分类'

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
  favorite_cnt: number
  created_at: string
}

export interface FavoriteGroupListData {
  groups: FavoriteGroup[]
  uncategorized_cnt: number
  total_favorite_cnt: number
}
