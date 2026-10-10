import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { setSessionExpiredHandler } from '@/lib/api';
import { authStorage } from '@/lib/auth-storage';
import type { RoleCode } from '@/types/api';
import { authApi } from '../api';
import type { LoginInput, User } from '../types';

type Status = 'loading' | 'authenticated' | 'guest';

interface AuthContextValue {
  status: Status;
  user: User | null;
  login: (input: LoginInput) => Promise<User>;
  logout: () => Promise<void>;
  hasRole: (...roles: RoleCode[]) => boolean;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient();
  const [user, setUser] = useState<User | null>(null);
  const [status, setStatus] = useState<Status>(() => (authStorage.get() ? 'loading' : 'guest'));

  const reset = useCallback(() => {
    authStorage.clear();
    queryClient.clear();
    setUser(null);
    setStatus('guest');
  }, [queryClient]);

  // Restore the session on page load.
  useEffect(() => {
    setSessionExpiredHandler(reset);
    if (!authStorage.get()) return;
    authApi
      .me()
      .then((u) => {
        setUser(u);
        setStatus('authenticated');
      })
      .catch(reset);
  }, [reset]);

  const login = useCallback(async (input: LoginInput) => {
    const res = await authApi.login(input);
    authStorage.set({ accessToken: res.access_token, refreshToken: res.refresh_token });
    setUser(res.user);
    setStatus('authenticated');
    return res.user;
  }, []);

  const logout = useCallback(async () => {
    const tokens = authStorage.get();
    try {
      if (tokens) await authApi.logout(tokens.refreshToken);
    } finally {
      reset();
    }
  }, [reset]);

  const value = useMemo<AuthContextValue>(
    () => ({
      status,
      user,
      login,
      logout,
      hasRole: (...roles) => !!user && roles.includes(user.role.code),
    }),
    [status, user, login, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used inside <AuthProvider>');
  return ctx;
}
