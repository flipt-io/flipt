type WellProps = {
  className?: string;
  children: React.ReactNode;
};

export default function Well(props: WellProps) {
  const { children, className } = props;

  return (
    <div className={`overflow-hidden rounded-lg border dark:border-gray-700 ${className}`}>
      <div className="flex flex-col items-center text-center p-8 space-y-6">
        {children}
      </div>
    </div>
  );
}
