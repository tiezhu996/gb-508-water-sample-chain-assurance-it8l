
import { CommonModule } from '@angular/common'; import { Component, Input } from '@angular/core'; import type { DomainRecord } from '../../types/domain'; import { StatusBadgeComponent } from './status-badge.component';
@Component({ selector: 'app-chain-timeline', standalone: true, imports: [CommonModule, StatusBadgeComponent], template: `<div *ngIf="records.length; else empty" class="evidence-strip"><article *ngFor="let item of records.slice(0, 4)"><strong>{{ item.code }}</strong><span>{{ item.name }}</span><app-status-badge [status]="item.status"/></article></div><ng-template #empty><div class="empty">暂无业务证据</div></ng-template>` })
export class ChainTimelineComponent { @Input() records: DomainRecord[] = []; }
