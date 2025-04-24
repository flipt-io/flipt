import { ChevronDownIcon, FlagIcon } from 'lucide-react';
import { useState } from 'react';
import { useNavigate } from 'react-router';

import { Badge } from '~/components/Badge';
import { Button } from '~/components/Button';
import { JsonEditor } from '~/components/json/JsonEditor';

import { FlagType, IFlag } from '~/types/Flag';

import { cls } from '~/utils/helpers';

import { FlagTypeBadge } from './FlagTypeBadge';

export function FlagDetails({ item }: { item: IFlag }) {
  if (item.type === FlagType.BOOLEAN) {
    // For boolean flags, show the actual value (true/false)
    return (
      <div className="flex items-center gap-2">
        <Badge variant={item.enabled ? 'enabled' : 'destructive'}>
          {item.enabled ? 'True' : 'False'}
        </Badge>
      </div>
    );
  }

  // For variant flags, show enabled/disabled state
  return (
    <div className="flex items-center gap-2">
      <Badge variant={item.enabled ? 'outline' : 'muted'}>
        {item.enabled ? 'Enabled' : 'Disabled'}
      </Badge>
    </div>
  );
}

export function FlagDetailPanel({
  flag,
  namespace
}: {
  flag: IFlag;
  namespace: string;
}) {
  const navigate = useNavigate();
  const [isMetadataExpanded, setIsMetadataExpanded] = useState(true);

  return (
    <div className="space-y-6 rounded-lg border p-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">{flag.name}</h2>
          <code className="text-sm text-muted-foreground font-mono">
            {flag.key}
          </code>
        </div>
        <div className="flex items-center gap-2">
          <FlagTypeBadge type={flag.type} />
          <FlagDetails item={flag} />
        </div>
      </div>

      {flag.description && (
        <div>
          <h3 className="text-sm font-medium text-secondary-foreground mb-1">
            Description
          </h3>
          <p className="text-sm text-secondary-foreground">
            {flag.description}
          </p>
        </div>
      )}

      <div>
        <button
          onClick={() => setIsMetadataExpanded(!isMetadataExpanded)}
          className="flex w-full items-center justify-between mb-1"
        >
          <h3 className="text-sm font-medium text-secondary-foreground">
            Metadata
          </h3>
          <ChevronDownIcon
            className={cls(
              'h-4 w-4 text-secondary-foreground transition-transform duration-200',
              {
                'transform rotate-180': isMetadataExpanded
              }
            )}
          />
        </button>
        {isMetadataExpanded && (
          <>
            {flag.metadata && Object.keys(flag.metadata).length > 0 ? (
              <div className="rounded-md border bg-muted/50">
                <JsonEditor
                  id="flag-metadata-viewer"
                  value={JSON.stringify(flag.metadata, null, 2)}
                  setValue={() => {}} // No-op since this is read-only
                  disabled={true}
                  strict={false}
                  height="20vh"
                />
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">
                No metadata is associated with this flag
              </p>
            )}
          </>
        )}
      </div>

      {/* Placeholder for evaluation analytics */}
      {/* {flag.enabled && (
          <div>
            <h3 className="text-sm font-medium text-secondary-foreground mb-1">Evaluation Analytics</h3>
            <div className="text-sm text-muted-foreground">Analytics coming soon...</div>
          </div>
        )} */}

      <div className="flex pt-4 justify-end">
        <Button
          variant="soft"
          onClick={() => navigate(`/namespaces/${namespace}/flags/${flag.key}`)}
        >
          Edit Flag
        </Button>
      </div>
    </div>
  );
}

export function EmptyFlagDetails() {
  return (
    <div className="flex h-full flex-col items-center justify-center text-center p-6">
      <FlagIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
      <h3 className="text-lg font-medium text-muted-foreground mb-2">
        Quick View Panel
      </h3>
      <div className="text-sm text-muted-foreground space-y-2">
        <p>Click the details button (→) to view flag information here.</p>
        <p>Click the flag row to edit it directly.</p>
      </div>
    </div>
  );
}
