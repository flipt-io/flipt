type WellProps = {
  className?: string;
  children: React.ReactNode;
};

export default function Well(props: WellProps) {
  const { children, className } = props;

  return (
    <div className={`bg-muted overflow-hidden rounded-lg ${className}`}>
      <div className="text-muted-foreground px-4 py-5 text-center text-sm sm:p-8">
        {children}
      </div>
    </div>
  );
}
