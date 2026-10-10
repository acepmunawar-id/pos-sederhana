const ACCESS = 'pos.access_token';
const REFRESH = 'pos.refresh_token';

export interface StoredTokens {
  accessToken: string;
  refreshToken: string;
}

export const authStorage = {
  get(): StoredTokens | null {
    const accessToken = localStorage.getItem(ACCESS);
    const refreshToken = localStorage.getItem(REFRESH);
    return accessToken && refreshToken ? { accessToken, refreshToken } : null;
  },
  set(tokens: StoredTokens) {
    localStorage.setItem(ACCESS, tokens.accessToken);
    localStorage.setItem(REFRESH, tokens.refreshToken);
  },
  clear() {
    localStorage.removeItem(ACCESS);
    localStorage.removeItem(REFRESH);
  },
};
