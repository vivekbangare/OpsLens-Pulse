import axios from "axios";

export const fetchTenants = async () => {
  const res = await axios.get("/api/admin/tenants");
  return res.data;
};

export const fetchUsers = async (tenantId: string) => {
  const res = await axios.get("/api/admin/users", {
    headers: { "X-Tenant-ID": tenantId },
  });
  return res.data;
};