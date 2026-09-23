
import { CommonModule } from '@angular/common';
import { Component, EventEmitter, Input, Output } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatInputModule } from '@angular/material/input';

@Component({
  selector: 'app-handover-dialog', standalone: true, imports: [CommonModule, FormsModule, MatButtonModule, MatInputModule],
  template: `<div *ngIf="open" class="modal-backdrop"><section class="modal" role="dialog" aria-label="保管交接">
    <h2>样本保管交接</h2>
    <p>交接仅同步当前保管人、交接时间与版本，不改变样本状态，并写入不可覆盖的审计日志。</p>
    <label class="handover-field">目标账号
      <input matInput name="target" [(ngModel)]="target" placeholder="启用且角色不低于 operator 的账号" maxlength="80"/>
    </label>
    <label class="handover-field">交接说明（可选）
      <input matInput name="reason" [(ngModel)]="reason" placeholder="例如：转交复核台" maxlength="500"/>
    </label>
    <small class="muted">目标账号须处于启用状态并具备 operator 及以上角色；并发交接只有一次会成功。</small>
    <footer><button mat-button (click)="cancel.emit()">取消</button><button mat-flat-button color="primary" [disabled]="!target.trim()" (click)="confirmHandover()">确认交接</button></footer>
  </section></div>`
})
export class HandoverDialogComponent {
  @Input() open = false;
  @Output() confirm = new EventEmitter<{ targetUsername: string; reason: string }>();
  @Output() cancel = new EventEmitter<void>();
  target = '';
  reason = '';
  confirmHandover() {
    const targetUsername = this.target.trim();
    if (!targetUsername) return;
    this.confirm.emit({ targetUsername, reason: this.reason.trim() });
  }
}
