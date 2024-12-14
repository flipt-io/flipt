import { useState } from 'react';
import { PlusIcon, TrashIcon } from '@heroicons/react/24/outline';
import * as Yup from 'yup';
import type { IFlagMetadata } from '~/types/Flag';
import { Button } from '~/components/ui/button';
import { Input } from '~/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '~/components/ui/select';
import { cls, stringAsKey } from '~/utils/helpers';
import { JsonEditor } from '~/components/json/JsonEditor';

const metadataValidationSchema = Yup.object({
  key: Yup.string().required('Key is required'),
  value: Yup.mixed().required('Value is required'),
  type: Yup.string().oneOf(['primitive', 'object', 'array']).required()
});

const VALUE_TYPES = [
  { id: 'primitive', name: 'Primitive' },
  { id: 'object', name: 'Object' },
  { id: 'array', name: 'Array' }
] as const;

export interface MetadataFormProps {
  metadata?: Record<string, any>;
  onChange: (metadata: Record<string, any>) => void;
  disabled?: boolean;
}

export function MetadataForm({
  metadata = {},
  onChange,
  disabled = false
}: MetadataFormProps): JSX.Element {
  const [entries, setEntries] = useState<IFlagMetadata[]>(
    objectToMetadataArray(metadata)
  );
  const [errors, setErrors] = useState<Record<number, Record<string, string>>>(
    {}
  );

  const validateEntry = async (
    entry: IFlagMetadata,
    index: number
  ): Promise<boolean> => {
    try {
      await metadataValidationSchema.validate(entry, { abortEarly: false });
      const newErrors = { ...errors };
      delete newErrors[index];
      setErrors(newErrors);
      return true;
    } catch (err) {
      if (err instanceof Yup.ValidationError) {
        const newErrors = { ...errors };
        newErrors[index] = err.inner.reduce(
          (acc, curr) => {
            if (curr.path) {
              acc[curr.path] = curr.message;
            }
            return acc;
          },
          {} as Record<string, string>
        );
        setErrors(newErrors);
      }
      return false;
    }
  };

  const handleMetadataChange = (newEntries: IFlagMetadata[]) => {
    setEntries(newEntries);
    onChange(metadataArrayToObject(newEntries));
  };

  const handleAdd = () => {
    const newEntries = [
      ...entries,
      {
        key: '',
        value: '',
        type: 'primitive',
        subtype: 'string',
        isNew: true
      } as IFlagMetadata
    ];
    setEntries(newEntries);
    handleMetadataChange(newEntries);
  };

  const handleRemove = (index: number) => {
    const newEntries = entries.filter((_, i) => i !== index);
    const newErrors = { ...errors };
    delete newErrors[index];
    setErrors(newErrors);
    setEntries(newEntries);
    handleMetadataChange(newEntries);
  };

  const handleChange = async (
    index: number,
    field: keyof IFlagMetadata,
    value: unknown
  ): Promise<void> => {
    const newEntries = [...entries];
    const entry = newEntries[index];

    if (field === 'key' && typeof value === 'string') {
      const formattedKey = stringAsKey(value);
      newEntries[index] = {
        ...entry,
        key: formattedKey
      };
    } else if (field === 'type' && typeof value === 'string') {
      const currentValue = entry.value;
      let newValue: any = null;

      if (value === 'object') {
        newValue =
          typeof currentValue === 'object' && !Array.isArray(currentValue)
            ? currentValue
            : {};
      } else if (value === 'array') {
        newValue = Array.isArray(currentValue) ? currentValue : [];
      } else {
        newValue =
          Array.isArray(currentValue) || typeof currentValue === 'object'
            ? ''
            : currentValue;
      }

      newEntries[index] = {
        ...entry,
        type: value as 'primitive' | 'object' | 'array',
        value: newValue
      };
    } else if (field === 'value') {
      let validValue: any;
      if (entry.type === 'primitive') {
        const primitiveType = getPrimitiveType(entry.value);
        if (typeof value === 'string') {
          if (primitiveType === 'string') {
            validValue = value;
          } else if (primitiveType === 'boolean') {
            validValue = value.toLowerCase() === 'true';
          } else if (primitiveType === 'number') {
            validValue = !isNaN(Number(value)) ? Number(value) : 0;
          }
        } else {
          validValue = value;
        }
      } else {
        validValue = typeof value === 'object' ? value : null;
      }

      newEntries[index] = {
        ...entry,
        value: validValue
      };
    }

    setEntries(newEntries);
    await validateEntry(newEntries[index], index);
    handleMetadataChange(newEntries);
  };

  const renderValueInput = (entry: IFlagMetadata, index: number) => {
    const error = errors[index]?.value;

    switch (entry.type) {
      case 'array':
        const arrayValue = Array.isArray(entry.value)
          ? JSON.stringify(entry.value)
          : '[]';
        return (
          <JsonEditor
            id={`metadata-value-${index}`}
            value={arrayValue}
            setValue={(v) => {
              try {
                const parsed = JSON.parse(v);
                if (Array.isArray(parsed)) {
                  handleChange(index, 'value', parsed);
                }
              } catch (e) {
                // Handle invalid JSON
              }
            }}
            disabled={disabled || !entry.isNew}
            strict={false}
            height="20vh"
          />
        );
      case 'object':
        const value =
          typeof entry.value === 'object'
            ? JSON.stringify(entry.value)
            : entry.value;
        return (
          <JsonEditor
            id={`metadata-value-${index}`}
            value={value}
            setValue={(v) => {
              try {
                const parsed = JSON.parse(v);
                if (typeof parsed === 'object' && !Array.isArray(parsed)) {
                  handleChange(index, 'value', parsed);
                }
              } catch (e) {
                // Handle invalid JSON
              }
            }}
            disabled={disabled || !entry.isNew}
            height="20vh"
          />
        );
      case 'primitive':
        return (
          <div className="flex gap-2">
            <Select
              value={entry.subtype || 'string'}
              onValueChange={(type) =>
                handlePrimitiveTypeChange(
                  index,
                  type as 'string' | 'number' | 'boolean'
                )
              }
              disabled={disabled || !entry.isNew}
            >
              <SelectTrigger className="w-24">
                <SelectValue placeholder="Type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="string">String</SelectItem>
                <SelectItem value="number">Number</SelectItem>
                <SelectItem value="boolean">Boolean</SelectItem>
              </SelectContent>
            </Select>
            {entry.subtype === 'boolean' ? (
              <Select
                value={String(entry.value)}
                onValueChange={(v) =>
                  handleChange(index, 'value', v === 'true')
                }
                disabled={disabled || !entry.isNew}
              >
                <SelectTrigger className="flex-1 disabled:opacity-75">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="true">True</SelectItem>
                  <SelectItem value="false">False</SelectItem>
                </SelectContent>
              </Select>
            ) : (
              <Input
                type={entry.subtype === 'number' ? 'number' : 'text'}
                value={entry.value?.toString() ?? ''}
                onChange={(e) => handleChange(index, 'value', e.target.value)}
                className={cls(
                  'flex-1 disabled:opacity-75',
                  error ? 'border-red-500' : ''
                )}
                aria-invalid={!!error}
                aria-errormessage={`value-error-${index}`}
                disabled={disabled || !entry.isNew}
                data-testid={`metadata-value-${index}`}
              />
            )}
          </div>
        );
    }
  };

  // Helper function to determine primitive type
  const getPrimitiveType = (value: any): 'string' | 'number' | 'boolean' => {
    switch (typeof value) {
      case 'number':
        return 'number';
      case 'boolean':
        return 'boolean';
      default:
        return 'string';
    }
  };

  // Handler for primitive type changes
  const handlePrimitiveTypeChange = (
    index: number,
    type: 'string' | 'number' | 'boolean'
  ) => {
    const entry = entries[index];
    let newValue: any;

    switch (type) {
      case 'number':
        newValue = !isNaN(Number(entry.value)) ? Number(entry.value) : 0;
        break;
      case 'boolean':
        newValue = Boolean(entry.value);
        break;
      case 'string':
        newValue = String(entry.value);
        break;
      default:
        newValue = entry.value;
    }

    const newEntries = [...entries];
    newEntries[index] = {
      ...entry,
      subtype: type,
      value: newValue
    };

    setEntries(newEntries);
    handleMetadataChange(newEntries);
  };

  return (
    <div className="space-y-4">
      {entries.map((entry, index) => {
        const keyError = errors[index]?.key;
        const typeError = errors[index]?.type;

        return (
          <div key={index} className="flex items-start gap-4">
            <div>
              <Input
                type="text"
                value={entry.key}
                onChange={(e) => handleChange(index, 'key', e.target.value)}
                placeholder="Key"
                className={cls(
                  'w-48 disabled:opacity-75',
                  keyError ? 'border-red-500' : ''
                )}
                aria-invalid={!!keyError}
                aria-errormessage={`key-error-${index}`}
                disabled={disabled || !entry.isNew}
                data-testid={`metadata-key-${index}`}
              />
              {keyError && (
                <p
                  className="mt-1 text-sm text-red-500"
                  id={`key-error-${index}`}
                >
                  {keyError}
                </p>
              )}
            </div>

            <div className="w-32">
              <Select
                value={entry.type}
                onValueChange={(value) => handleChange(index, 'type', value)}
                disabled={disabled || !entry.isNew}
              >
                <SelectTrigger
                  className={cls('w-full', typeError ? 'border-red-500' : '')}
                  aria-invalid={!!typeError}
                  aria-errormessage={`type-error-${index}`}
                  data-testid={`metadata-type-${index}`}
                >
                  <SelectValue placeholder="Type" />
                </SelectTrigger>
                <SelectContent>
                  {VALUE_TYPES.map((type) => (
                    <SelectItem key={type.id} value={type.id}>
                      {type.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {typeError && (
                <p
                  className="mt-1 text-sm text-red-500"
                  id={`type-error-${index}`}
                >
                  {typeError}
                </p>
              )}
            </div>

            <div className="flex-1">{renderValueInput(entry, index)}</div>

            <Button
              type="button"
              variant="ghost"
              size="icon"
              onClick={() => handleRemove(index)}
              disabled={disabled}
              aria-label="Remove metadata entry"
            >
              <TrashIcon className="h-5 w-5 text-gray-500" aria-hidden="true" />
            </Button>
          </div>
        );
      })}

      <Button
        type="button"
        variant="outline"
        onClick={handleAdd}
        disabled={disabled}
        className="w-full"
      >
        <PlusIcon className="mr-2 h-5 w-5" />
        Add Metadata
      </Button>
    </div>
  );
}

function objectToMetadataArray(obj: Record<string, any>): IFlagMetadata[] {
  return Object.entries(obj).map(([key, value]) => {
    if (value === null || value === undefined) {
      return {
        key,
        type: 'primitive',
        subtype: 'string',
        value: '',
        isNew: false
      };
    }

    if (Array.isArray(value)) {
      return {
        key,
        type: 'array',
        value,
        isNew: false
      };
    }

    if (typeof value === 'object' && value !== null) {
      return {
        key,
        type: 'object',
        value,
        isNew: false
      };
    }

    let subtype: 'string' | 'number' | 'boolean' = 'string';

    if (typeof value === 'string') {
      // Try parsing as number first
      if (!isNaN(Number(value)) && value.trim() !== '') {
        subtype = 'number';
      }
      // Try parsing as boolean if it's exactly 'true' or 'false'
      else if (
        value.toLowerCase() === 'true' ||
        value.toLowerCase() === 'false'
      ) {
        subtype = 'boolean';
      }
    } else if (typeof value === 'number') {
      subtype = 'number';
    } else if (typeof value === 'boolean') {
      subtype = 'boolean';
    }

    return {
      key,
      type: 'primitive',
      subtype,
      value,
      isNew: false
    };
  });
}

function metadataArrayToObject(metadata: IFlagMetadata[]): Record<string, any> {
  return metadata.reduce(
    (acc, { key, value }) => {
      if (key) {
        // Only include entries with non-empty keys
        acc[key] = value;
      }
      return acc;
    },
    {} as Record<string, any>
  );
}
