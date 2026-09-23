
import { Component } from '@angular/core'; import { EntityPageComponent } from '../components/entity-page.component'; import { ENTITY_CONFIGS } from '../types/status'; import { LabSampleStore } from '../stores/lab-sample.store';
@Component({ selector: 'app-lab-sample-page', standalone: true, imports: [EntityPageComponent], template: `<app-entity-page [config]="config" [store]="store"/>` })
export class LabSamplePage { readonly config = ENTITY_CONFIGS[1]; constructor(readonly store: LabSampleStore) {} }
