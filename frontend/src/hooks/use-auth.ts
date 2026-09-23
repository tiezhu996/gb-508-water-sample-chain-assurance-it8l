
import { signal } from '@angular/core'; import { login } from '../api/auth'; import { clearSession, getSession, saveSession } from '../api/client'; import type { UserSession } from '../types/domain';

const ROLE_RANK: Record<string, number> = { viewer: 1, operator: 2, reviewer: 3, admin: 4 };
export function hasMinimumRole(role: string | undefined, minimum: string): boolean { return (ROLE_RANK[role || ''] || 0) >= (ROLE_RANK[minimum] || Number.MAX_SAFE_INTEGER); }

export function createAuthState() {
  const stored = getSession(); const session = signal<UserSession | null>(stored); const loading = signal(!stored); let inFlight: Promise<UserSession> | null = null;
  const authenticate = (force = false): Promise<UserSession> => {
    if (!force && session()) { loading.set(false); return Promise.resolve(session()!); }
    if (inFlight) return inFlight;
    loading.set(true);
    inFlight = login().then((next) => { saveSession(next); session.set(next); return next; }).finally(() => { loading.set(false); inFlight = null; });
    return inFlight;
  };
  const logout = async () => { clearSession(); session.set(null); return authenticate(true); };
  return { session, loading, authenticate, logout };
}

export const authState = createAuthState();
