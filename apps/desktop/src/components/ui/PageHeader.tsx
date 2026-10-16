interface Props { title: string; actions?: React.ReactNode }
export const PageHeader = ({ title, actions }: Props) => (
  <div className="flex items-center justify-between mb-8">
    <h1 className="text-[1.5rem] font-semibold tracking-tight text-ink">{title}</h1>
    {actions && <div>{actions}</div>}
  </div>
);
