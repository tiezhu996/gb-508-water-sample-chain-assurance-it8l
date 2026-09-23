import { CommonModule } from '@angular/common';
import { Component, Input } from '@angular/core';
import type { DomainRecord } from '../../types/domain';
import { StatusBadgeComponent } from './status-badge.component';

@Component({
  selector: 'app-chain-badge', standalone: true, imports: [CommonModule, StatusBadgeComponent],
  template: `<section class="context-panel" aria-label="样本链路摘要"><header><strong>样本链路</strong><span>{{ records.length }} 条可追踪记录</span></header><div class="context-items"><article *ngFor="let item of records.slice(0, 4)"><div><strong>{{ item.relatedCode || item.code }}</strong><small>{{ item.code }} · {{ item.facility }}</small></div><app-status-badge [status]="item.status"/></article></div></section>`
})
export class ChainBadgeComponent { @Input() records: DomainRecord[] = []; }
