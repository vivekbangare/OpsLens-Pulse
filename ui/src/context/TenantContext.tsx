import { createContext, useContext, useEffect, useState } from "react";
import { useAuth } from "@/features/auth/AuthContext";

export interface Tenant {
  id: string;
  name: string;
  slug: string;
  plan: string;
}

interface TenantContextType {
  currentTenant: Tenant | null;
  tenants: Tenant[];
  setCurrentTenant: (tenant: Tenant) => void;
  setTenants: (tenants: Tenant[]) => void;
}

const TenantContext = createContext<TenantContextType | undefined>(undefined);

export const TenantProvider = ({ children }: { children: React.ReactNode }) => {
  const { user } = useAuth();
  const [currentTenant, setCurrentTenantState] = useState<Tenant | null>(null);
  const [tenants, setTenants] = useState<Tenant[]>([]);
  useEffect(() => {
    if (user?.tenants?.length) {
      setTenants(user.tenants);

      const defaultTenant =
        user.tenants.find(t => t.id === user.currentTenantId) ||
        user.tenants[0];

      setCurrentTenant(defaultTenant);
    }
  }, [user]);

  useEffect(() => {
    const saved = localStorage.getItem("currentTenant");
    if (saved) {
      setCurrentTenantState(JSON.parse(saved));
    }
  }, []);

  const setCurrentTenant = (tenant: Tenant) => {
    setCurrentTenantState(tenant);
    localStorage.setItem("currentTenant", JSON.stringify(tenant));
  };

  return (
    <TenantContext.Provider
      value={{
        currentTenant,
        tenants,
        setCurrentTenant,
        setTenants,
      }}
    >
      {children}
    </TenantContext.Provider>
  );
};

export const useTenant = () => {
  const context = useContext(TenantContext);
  if (!context) {
    throw new Error("useTenant must be used inside TenantProvider");
  }
  return context;
};