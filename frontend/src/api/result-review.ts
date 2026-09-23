
import { request } from './client';
import type { DomainRecord } from '../types/domain';

export async function listResultReview(page = 1, pageSize = 20, search = '') {
  return request<DomainRecord[]>(`/reviews?page=${page}&pageSize=${pageSize}&search=${encodeURIComponent(search)}`);
}
export async function createResultReview(input: Partial<DomainRecord>) {
  return request<DomainRecord>('/reviews', { method: 'POST', body: JSON.stringify(input) });
}
export async function transitionResultReview(id: number, status: string, expectedVersion: number, reason: string) {
  return request<DomainRecord>(`/reviews/${id}/transition`, {
    method: 'POST', body: JSON.stringify({ status, expectedVersion, reason }),
  });
}
