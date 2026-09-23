import type { EntityConfig } from './domain';

export type SampleState = 'received' | 'accepted' | 'testing' | 'hold' | 'disposed';
export const ALL_SAMPLE_STATE: readonly SampleState[] = ['received', 'accepted', 'testing', 'hold', 'disposed'];
export type ReviewState = 'draft' | 'peer_review' | 'signed' | 'rejected';
export const ALL_REVIEW_STATE: readonly ReviewState[] = ['draft', 'peer_review', 'signed', 'rejected'];

export const ENTITY_CONFIGS: readonly EntityConfig[] = [
  { key: 'samplingBatch', path: 'sampling-batches', label: '采样批次', statuses: ['planned', 'collecting', 'received', 'closed'] as const },
  { key: 'labSample', path: 'samples', label: '实验室样本', statuses: ['received', 'accepted', 'testing', 'hold', 'disposed'] as const },
  { key: 'assayMethod', path: 'methods', label: '检测方法', statuses: ['draft', 'validated', 'active', 'retired'] as const },
  { key: 'resultReview', path: 'reviews', label: '结果复核', statuses: ['draft', 'peer_review', 'signed', 'rejected'] as const }
];
