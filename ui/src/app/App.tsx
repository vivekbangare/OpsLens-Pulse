import AppRoutes from "./AppRoutes";
import { TenantProvider } from "@/context/TenantContext";

export default function App() {
  return (
    <TenantProvider>
      <AppRoutes />
    </TenantProvider>
  );
}