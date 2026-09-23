
import { request } from './client';
import type { DomainRecord } from '../types/domain';

export async function listSamplingBatch(page = 1, pageSize = 20, search = '') {
  return request<DomainRecord[]>(`/sampling-batches?page=${page}&pageSize=${pageSize}&search=${encodeURIComponent(search)}`);
}
export async function createSamplingBatch(input: Partial<DomainRecord>) {
  return request<DomainRecord>('/sampling-batches', { method: 'POST', body: JSON.stringify(input) });
}
export async function transitionSamplingBatch(id: number, status: string, expectedVersion: number, reason: string) {
  return request<DomainRecord>(`/sampling-batches/${id}/transition`, {
    method: 'POST', body: JSON.stringify({ status, expectedVersion, reason }),
  });
}
