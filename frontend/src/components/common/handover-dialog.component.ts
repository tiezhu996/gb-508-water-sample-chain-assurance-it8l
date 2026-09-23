
import { CommonModule } from '@angular/common'; import { Component, EventEmitter, Input, OnChanges, Output } from '@angular/core'; import { FormsModule } from '@angular/forms'; import { MatButtonModule } from '@angular/material/button'; import { MatInputModule } from '@angular/material/input'; import { formatDate } from '../../utils/format'; import type { DomainRecord } from '../../types/domain';

@Component({
  selector: 'app-handover-dialog', standalone: true, imports: [CommonModule, FormsModule, MatButtonModule, MatInputModule],
  template: `<div *ngIf="open && item" class="modal-backdrop"><section class="modal" role="dialog" aria-label="样本保管交接"><h2>样本保管交接</h2>
    <p class="muted">{{ item.code }} · {{ item.name }}</p>
    <dl class="handover-summary"><dt>当前保管人</dt><dd>{{ item.custodian || '-' }}</dd><dt>最近交接时间</dt><dd>{{ formatDate(item.handoverAt || '') }}</dd></dl>
    <label class="handover-field" for="handover-target">目标账号</label>
    <input id="handover-target" matInput [(ngModel)]="targetUsername" placeholder="输入启用且具备 operator 以上角色的账号" aria-label="目标账号"/>
    <label class="handover-field" for="handover-remark">交接备注（可选）</label>
    <input id="handover-remark" matInput [(ngModel)]="remark" maxlength="500" placeholder="例如：随批次移交二号实验室" aria-label="交接备注"/>
    <p class="handover-hint">目标账号必须启用且具备 operator 以上角色；样本已处置、目标为当前保管人或版本过期时将被拒绝。</p>
    <div *ngIf="error" class="alert">{{ error }}</div>
    <footer><button mat-button (click)="cancel.emit()" [disabled]="submitting">取消</button><button mat-flat-button color="primary" (click)="confirm()" [disabled]="submitting || !targetUsername.trim()">{{ submitting ? '提交中…' : '确认交接' }}</button></footer>
  </section></div>`
})
export class HandoverDialogComponent implements OnChanges {
  @Input() open = false;
  @Input() item: DomainRecord | null = null;
  @Input() error = '';
  @Input() submitting = false;
  @Output() confirm = new EventEmitter<{ targetUsername: string; remark: string }>();
  @Output() cancel = new EventEmitter<void>();
  targetUsername = '';
  remark = '';
  readonly formatDate = formatDate;
  ngOnChanges() { if (!this.open) { this.targetUsername = ''; this.remark = ''; } }
}
