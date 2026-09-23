
import { Component, Input } from '@angular/core';
@Component({ selector: 'app-empty-state', standalone: true, template: `<div class="empty-state"><strong>{{ title }}</strong><span>{{ detail }}</span></div>` })
export class EmptyStateComponent { @Input() title = '暂无记录'; @Input() detail = '调整筛选条件后重试'; }
