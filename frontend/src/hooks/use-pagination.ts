
import { computed, signal } from '@angular/core'; export function createPagination(total: () => number, initialPageSize = 20) { const page = signal(1); const pageSize = signal(initialPageSize); const pages = computed(() => Math.max(1, Math.ceil(total() / pageSize()))); return { page, pageSize, pages }; }
