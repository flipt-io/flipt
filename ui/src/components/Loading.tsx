type LoadingProps = {
  isPrimary?: boolean;
};

export default function Loading(props: LoadingProps) {
  const { isPrimary } = props;

  return (
    <div className="flex items-center justify-center">
      <div
        className={`h-5 w-5 ${
          isPrimary ? 'border-white-300' : 'border-violet-300'
        } animate-spin rounded-full border-b-2`}
      ></div>
    </div>
  );
}
