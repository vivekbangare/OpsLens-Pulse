interface Column<T> {
  header: string;
  accessor: keyof T;
}

interface Props<T> {
  columns: Column<T>[];
  data: T[];
}

export default function AdminTable<T extends object>({
  columns,
  data,
}: Props<T>) {
  return (
    <table className="adminTable">
      <thead>
        <tr>
          {columns.map((col, i) => (
            <th key={i}>{col.header}</th>
          ))}
        </tr>
      </thead>

      <tbody>
        {data.length === 0 && (
          <tr>
            <td colSpan={columns.length}>
              <div className="emptyState">
                No data available
              </div>
            </td>
          </tr>
        )}

        {data.map((row, i) => (
          <tr key={i}>
            {columns.map((col, j) => (
              <td key={j}>
                {String(row[col.accessor] ?? "—")}
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );
}