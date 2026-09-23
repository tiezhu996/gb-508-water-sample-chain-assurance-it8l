
import { Component } from '@angular/core'; import { EntityPageComponent } from '../components/entity-page.component'; import { ENTITY_CONFIGS } from '../types/status'; import { AssayMethodStore } from '../stores/assay-method.store';
@Component({ selector: 'app-assay-method-page', standalone: true, imports: [EntityPageComponent], template: `<app-entity-page [config]="config" [store]="store"/>` })
export class AssayMethodPage { readonly config = ENTITY_CONFIGS[2]; constructor(readonly store: AssayMethodStore) {} }
