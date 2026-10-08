interface Props { title: string; actions?: React.ReactNode }
export const PageHeader = ({ title, actions }: Props) => (
  <div className="flex items-center justify-between mb-6">
    <h1 className="text-2xl font-semibold">{title}</h1>
    {actions && <div>{actions}</div>}
  </div>
);
