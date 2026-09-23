
import { Component } from '@angular/core'; import { EntityPageComponent } from '../components/entity-page.component'; import { ENTITY_CONFIGS } from '../types/status'; import { SamplingBatchStore } from '../stores/sampling-batch.store';
@Component({ selector: 'app-sampling-batch-page', standalone: true, imports: [EntityPageComponent], template: `<app-entity-page [config]="config" [store]="store"/>` })
export class SamplingBatchPage { readonly config = ENTITY_CONFIGS[0]; constructor(readonly store: SamplingBatchStore) {} }
