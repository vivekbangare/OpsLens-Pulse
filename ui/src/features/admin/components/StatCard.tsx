interface StatCardProps {
  title: string;
  count: number;
  subtitle: string;
  onClick?: () => void;
}

export default function StatCard({
  title,
  count,
  subtitle,
  onClick,
}: StatCardProps) {
  return (
    <div className="stat-card" onClick={onClick}>
      <h3>{title}</h3>
      <h1>{count}</h1>
      <p>{subtitle}</p>
    </div>
  );
}