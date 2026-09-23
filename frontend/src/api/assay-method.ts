
import { request } from './client';
import type { DomainRecord } from '../types/domain';

export async function listAssayMethod(page = 1, pageSize = 20, search = '') {
  return request<DomainRecord[]>(`/methods?page=${page}&pageSize=${pageSize}&search=${encodeURIComponent(search)}`);
}
export async function createAssayMethod(input: Partial<DomainRecord>) {
  return request<DomainRecord>('/methods', { method: 'POST', body: JSON.stringify(input) });
}
export async function transitionAssayMethod(id: number, status: string, expectedVersion: number, reason: string) {
  return request<DomainRecord>(`/methods/${id}/transition`, {
    method: 'POST', body: JSON.stringify({ status, expectedVersion, reason }),
  });
}
