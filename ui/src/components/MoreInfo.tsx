type MoreInfoProps = {
  className?: string;
  href: string;
  children: React.ReactNode;
};

export default function MoreInfo(props: MoreInfoProps) {
  const { className, href, children } = props;

  return (
    <div className={`${className} flex text-xs tracking-tight`}>
      <a
        href={href}
        target="_blank"
        rel="noreferrer"
        className="group inline-flex items-center text-gray-500 underline underline-offset-4 hover:text-gray-600"
      >
        {children}
      </a>
    </div>
  );
}
