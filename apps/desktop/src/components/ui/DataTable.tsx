export const DataTable = ({ children }: { children: React.ReactNode }) => (
  <div className="overflow-x-auto">
    <table className="w-full text-sm">{children}</table>
  </div>
);
