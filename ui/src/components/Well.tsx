type WellProps = {
  className?: string;
  children: React.ReactNode;
};

export default function Well(props: WellProps) {
  const { children, className } = props;

  return (
    <div className={`overflow-hidden rounded-lg bg-gray-50 ${className}`}>
      <div className="text-muted-foreground px-4 py-5 text-center text-sm sm:p-8">
        {children}
      </div>
    </div>
  );
}
