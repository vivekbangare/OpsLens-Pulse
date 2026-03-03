import { Route } from "react-router-dom";
import AdminLayout from "./AdminLayout";
import UsersPage from "./pages/UsersPage";
import GroupsPage from "./pages/GroupsPage";
import TenantsPage from "./pages/TenantsPage";

export const AdminRoutes = (
  <Route path="admin" element={<AdminLayout />}>
    <Route index element={<UsersPage />} />
    <Route path="users" element={<UsersPage />} />
    <Route path="groups" element={<GroupsPage />} />
    <Route path="tenants" element={<TenantsPage />} />
  </Route>
);