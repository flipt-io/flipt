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
        className="group text-muted-foreground hover:text-secondary-foreground inline-flex items-center underline underline-offset-4"
      >
        {children}
      </a>
    </div>
  );
}
