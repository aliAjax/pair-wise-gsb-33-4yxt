import request from '@/utils/request'
import type { Favorite, FavoriteGroup, FavoriteTargetType } from '@/constants/favorite'

export function listFavorites(params?: { target_type?: FavoriteTargetType | ''; group_id?: number }) {
  return request.get<never, Favorite[]>('/favorites', { params })
}

export function addFavorite(targetType: FavoriteTargetType, targetId: number, groupId = 0) {
  return request.post<never, Favorite>('/favorites', {
    target_type: targetType,
    target_id: targetId,
    group_id: groupId,
  })
}

export function removeFavorite(targetType: FavoriteTargetType, targetId: number) {
  return request.delete<never, { removed: boolean }>(`/favorites/${targetType}/${targetId}`)
}

export function moveFavorite(favoriteId: number, groupId: number) {
  return request.put<never, Favorite>(`/favorites/${favoriteId}/group`, { group_id: groupId })
}

export function listFavoriteGroups() {
  return request.get<never, FavoriteGroup[]>('/favorite-groups')
}

export function createFavoriteGroup(name: string) {
  return request.post<never, FavoriteGroup>('/favorite-groups', { name })
}

export function renameFavoriteGroup(id: number, name: string) {
  return request.put<never, FavoriteGroup>(`/favorite-groups/${id}`, { name })
}

export function deleteFavoriteGroup(id: number) {
  return request.delete<never, { deleted: boolean; detached: number }>(`/favorite-groups/${id}`)
}
