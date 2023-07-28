type WellProps = {
  className?: string;
  children: React.ReactNode;
};

export default function Well(props: WellProps) {
  const { children, className } = props;

  return (
    <div className={`bg-gray-50 overflow-hidden rounded-lg ${className}`}>
      <div className="px-4 py-5 sm:p-8">{children}</div>
    </div>
  );
}
