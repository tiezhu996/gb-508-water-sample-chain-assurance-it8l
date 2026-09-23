
import { Component } from '@angular/core'; import { EntityPageComponent } from '../components/entity-page.component'; import { ENTITY_CONFIGS } from '../types/status'; import { ResultReviewStore } from '../stores/result-review.store';
@Component({ selector: 'app-result-review-page', standalone: true, imports: [EntityPageComponent], template: `<app-entity-page [config]="config" [store]="store"/>` })
export class ResultReviewPage { readonly config = ENTITY_CONFIGS[3]; constructor(readonly store: ResultReviewStore) {} }
