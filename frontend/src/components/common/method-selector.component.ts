import { CommonModule } from '@angular/common';
import { Component, Input } from '@angular/core';
import type { DomainRecord } from '../../types/domain';

@Component({
  selector: 'app-method-selector', standalone: true, imports: [CommonModule],
  template: `<section class="method-selector"><label for="method-version">检测方法版本</label><select id="method-version" aria-label="检测方法版本" [value]="selected" (change)="selected = valueOf($event)"><option value="">选择适用方法</option><option *ngFor="let item of records" [value]="codeOf(item)">{{ codeOf(item) }} · {{ item.status }}</option></select></section>`
})
export class MethodSelectorComponent { @Input() records: DomainRecord[] = []; @Input() useRelatedCode = false; selected = ''; codeOf(item: DomainRecord) { return this.useRelatedCode ? item.relatedCode || item.code : item.code; } valueOf(event: Event) { return (event.target as HTMLSelectElement).value; } }
