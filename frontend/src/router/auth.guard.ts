import type { CanActivateFn } from '@angular/router';
import { authState } from '../hooks/use-auth';

export const authGuard: CanActivateFn = async () => {
  try { await authState.authenticate(); return true; }
  catch { return false; }
};
