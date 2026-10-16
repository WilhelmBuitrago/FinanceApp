import React from 'react';

interface PageContainerProps {
  children: React.ReactNode;
}

export default function PageContainer({ children }: PageContainerProps) {
  return (
    <main className="flex-1 overflow-auto bg-bg">
      <div className="mx-auto h-full w-full max-w-[1440px] px-8 py-8">{children}</div>
    </main>
  );
}
