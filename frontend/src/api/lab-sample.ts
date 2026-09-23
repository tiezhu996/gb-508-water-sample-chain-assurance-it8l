
import { request } from './client';
import type { DomainRecord } from '../types/domain';

export async function listLabSample(page = 1, pageSize = 20, search = '') {
  return request<DomainRecord[]>(`/samples?page=${page}&pageSize=${pageSize}&search=${encodeURIComponent(search)}`);
}
export async function createLabSample(input: Partial<DomainRecord>) {
  return request<DomainRecord>('/samples', { method: 'POST', body: JSON.stringify(input) });
}
export async function transitionLabSample(id: number, status: string, expectedVersion: number, reason: string) {
  return request<DomainRecord>(`/samples/${id}/transition`, {
    method: 'POST', body: JSON.stringify({ status, expectedVersion, reason }),
  });
}
