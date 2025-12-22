// Token management
export const getToken = (): string | null => {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem('token');
  };
  
  export const setToken = (token: string): void => {
    if (typeof window === 'undefined') return;
    localStorage.setItem('token', token);
  };
  
  export const removeToken = (): void => {
    if (typeof window === 'undefined') return;
    localStorage.removeItem('token');
  };
  
  export const isAuthenticated = (): boolean => {
    return !!getToken();
  };
  
  // Redirect to login if not authenticated
  export const requireAuth = (): boolean => {
    if (typeof window === 'undefined') return false;
    
    if (!isAuthenticated()) {
      window.location.href = '/';
      return false;
    }
    
    return true;
  };
  
  // Logout and redirect
  export const logout = (): void => {
    removeToken();
    window.location.href = '/';
  };