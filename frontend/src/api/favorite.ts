import request from '@/utils/request'
import type { Favorite, FavoriteGroupListData, FavoriteTargetType } from '@/constants/favorite'

export interface FavoriteListParams {
  target_type?: FavoriteTargetType | ''
  group_id?: number
  uncategorized?: boolean
}

export function listFavorites(params: FavoriteListParams = {}) {
  const query: Record<string, string | number> = {}
  if (params.target_type) query.target_type = params.target_type
  if (params.group_id) query.group_id = params.group_id
  if (params.uncategorized) query.uncategorized = 'true'
  return request.get<never, Favorite[]>('/favorites', { params: query })
}

export function addFavorite(targetType: FavoriteTargetType, targetId: number, groupId?: number | null) {
  return request.post<never, Favorite>('/favorites', {
    target_type: targetType,
    target_id: targetId,
    group_id: groupId ?? null,
  })
}

// groupId null/0 moves the favorite back to "未分类".
export function moveFavorite(targetType: FavoriteTargetType, targetId: number, groupId: number | null) {
  return request.put<never, Favorite>(`/favorites/${targetType}/${targetId}/group`, {
    group_id: groupId ?? null,
  })
}

export function removeFavorite(targetType: FavoriteTargetType, targetId: number) {
  return request.delete<never, { removed: boolean }>(`/favorites/${targetType}/${targetId}`)
}

export function listFavoriteGroups() {
  return request.get<never, FavoriteGroupListData>('/favorites/groups')
}

export function createFavoriteGroup(name: string) {
  return request.post<never, import('@/constants/favorite').FavoriteGroup>('/favorites/groups', { name })
}

export function deleteFavoriteGroup(id: number) {
  return request.delete<never, { removed: boolean }>(`/favorites/groups/${id}`)
}
