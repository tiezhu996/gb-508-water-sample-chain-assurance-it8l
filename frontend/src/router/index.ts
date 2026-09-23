import type { Routes } from '@angular/router';
import { SamplingBatchPage } from '../pages/sampling-batch.page';
import { LabSamplePage } from '../pages/lab-sample.page';
import { AssayMethodPage } from '../pages/assay-method.page';
import { ResultReviewPage } from '../pages/result-review.page';
import { AuditPage } from '../pages/audit.page';
import { authGuard } from './auth.guard';
export const routes: Routes = [{ path: '', pathMatch: 'full', redirectTo: 'sampling-batches' }, { path: 'sampling-batches', component: SamplingBatchPage, canActivate: [authGuard] }, { path: 'samples', component: LabSamplePage, canActivate: [authGuard] }, { path: 'methods', component: AssayMethodPage, canActivate: [authGuard] }, { path: 'reviews', component: ResultReviewPage, canActivate: [authGuard] }, { path: 'audit', component: AuditPage, canActivate: [authGuard] }];
