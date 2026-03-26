import { ComponentType, ReactNode } from 'react';

interface PreferenceSectionProps {
  title: string;
  children: ReactNode;
}

export function PreferenceSection({ title, children }: PreferenceSectionProps) {
  return (
    <div className="bg-sidebar rounded-lg shadow-sm border mb-6 p-6">
      <h3 className="text-lg font-medium text-foreground mb-4">{title}</h3>
      {children}
    </div>
  );
}

interface PreferenceLabelProps {
  icon: ComponentType<{ className?: string }>;
  label: string;
  description: string;
  children: ReactNode;
}

export function PreferenceLabel({
  icon: Icon,
  label,
  description,
  children
}: PreferenceLabelProps) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between">
      <div className="mb-3 sm:mb-0">
        <div className="flex items-center">
          <Icon className="h-5 w-5 text-brand mr-2" />
          <span className="text-sm font-medium text-foreground/80">
            {label}
          </span>
        </div>
        <p className="mt-1 text-xs text-muted-foreground">{description}</p>
      </div>
      {children}
    </div>
  );
}
