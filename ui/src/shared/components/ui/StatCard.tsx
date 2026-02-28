import "./statcard.css"

interface Props {
  label: string
  value: string | number
}

export default function StatCard({ label, value }: Props) {
  return (
    <div className="statCard">
      <div className="statLabel">{label}</div>
      <div className="statValue">{value}</div>
    </div>
  )
}