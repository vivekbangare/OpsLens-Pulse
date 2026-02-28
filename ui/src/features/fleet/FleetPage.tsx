import StatCard from "../../shared/components/ui/StatCard"

export default function FleetOverview() {
  return (
    <div>
      <h2>Fleet Overview</h2>
      <p className="subtitle">
        Real-time infrastructure health summary
      </p>

      <div className="grid">
        <StatCard label="Total Hosts" value={12} />
        <StatCard label="Alive" value={10} />
        <StatCard label="Dead" value={2} />
        <StatCard label="Fleet Health" value="83%" />
      </div>
    </div>
  )
}