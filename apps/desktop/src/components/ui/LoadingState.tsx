export const LoadingState = ({ message = "Cargando..." }: { message?: string }) => (
  <div className="flex items-center justify-center py-12 text-sm text-muted-foreground">
    {message}
  </div>
);
