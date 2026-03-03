import AdminTable from "../components/AdminTable";

interface User {
  username: string;
  email: string;
  role: string;
  status: string;
}

export default function UsersPage() {
  const users: User[] = [
    { username: "john.doe", email: "john@email.com", role: "Admin", status: "Active" },
    { username: "jane.smith", email: "jane@email.com", role: "Member", status: "Active" },
  ];

  const columns = [
    { header: "Username", accessor: "username" },
    { header: "Email", accessor: "email" },
    { header: "Role", accessor: "role" },
    { header: "Status", accessor: "status" },
  ] as const;

  return (
    <div>
      <div className="adminToolbar">
        <button className="btn primary">+ Create User</button>
      </div>

      <AdminTable columns={columns as any} data={users} />
    </div>
  );
}