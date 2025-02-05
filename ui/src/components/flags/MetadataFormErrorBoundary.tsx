import React from 'react';

import { JsonEditor } from '~/components/json/JsonEditor';

interface MetadataFormErrorBoundaryProps {
  children: React.ReactNode;
  metadata?: Record<string, any>;
  onChange: (metadata: Record<string, any>) => void;
  disabled?: boolean;
}

class MetadataFormErrorBoundary extends React.Component<
  MetadataFormErrorBoundaryProps,
  { hasError: boolean }
> {
  constructor(props: MetadataFormErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(_: Error) {
    return { hasError: true };
  }

  render() {
    if (this.state.hasError) {
      return (
        <JsonEditor
          id="metadata-fallback"
          value={JSON.stringify(this.props.metadata, null, 2)}
          setValue={(value) => {
            try {
              const parsed = JSON.parse(value);
              this.props.onChange(parsed);
            } catch (e) {
              console.warn('Invalid JSON:', e);
            }
          }}
          disabled={this.props.disabled}
          strict={false}
          height="20vh"
          data-testid="metadata-fallback"
        />
      );
    }

    return this.props.children;
  }
}

export default MetadataFormErrorBoundary;
