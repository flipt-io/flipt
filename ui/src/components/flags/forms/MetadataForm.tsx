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
import { stringAsKey } from '~/utils/helpers';

const metadataValidationSchema = Yup.object({
  key: Yup.string().required('Key is required'),
  value: Yup.mixed().required('Value is required'),
  type: Yup.string().oneOf(['string', 'boolean', 'number']).required()
});

const VALUE_TYPES = [
  { id: 'string', name: 'String' },
  { id: 'boolean', name: 'Boolean' },
  { id: 'number', name: 'Number' }
] as const;

export interface MetadataFormProps {
  metadata?: IFlagMetadata[];
  onChange: (metadata: IFlagMetadata[]) => void;
  disabled?: boolean;
}

export function MetadataForm({ metadata = [], onChange, disabled = false }: MetadataFormProps): JSX.Element {
  const [entries, setEntries] = useState<IFlagMetadata[]>(metadata);
  const [errors, setErrors] = useState<Record<number, Record<string, string>>>({});

  const validateEntry = async (entry: IFlagMetadata, index: number): Promise<boolean> => {
    try {
      await metadataValidationSchema.validate(entry, { abortEarly: false });
      const newErrors = { ...errors };
      delete newErrors[index];
      setErrors(newErrors);
      return true;
    } catch (err) {
      if (err instanceof Yup.ValidationError) {
        const newErrors = { ...errors };
        newErrors[index] = err.inner.reduce((acc, curr) => {
          if (curr.path) {
            acc[curr.path] = curr.message;
          }
          return acc;
        }, {} as Record<string, string>);
        setErrors(newErrors);
      }
      return false;
    }
  };

  const handleAdd = () => {
    const newEntries = [
      ...entries,
      { key: '', value: '', type: 'string' } as IFlagMetadata
    ];
    setEntries(newEntries);
    onChange(newEntries);
  };

  const handleRemove = (index: number) => {
    const newEntries = entries.filter((_, i) => i !== index);
    const newErrors = { ...errors };
    delete newErrors[index];
    setErrors(newErrors);
    setEntries(newEntries);
    onChange(newEntries);
  };

  const handleChange = async (index: number, field: keyof IFlagMetadata, value: unknown): Promise<void> => {
    const newEntries = [...entries];
    if (field === 'key' && typeof value === 'string') {
      value = stringAsKey(value);
    }
    if (field === 'type' && typeof value === 'string') {
      newEntries[index] = {
        ...newEntries[index],
        type: value as 'string' | 'boolean' | 'number',
        value: value === 'boolean' ? false : ''
      };
    } else {
      newEntries[index] = {
        ...newEntries[index],
        [field]: value
      };
    }
    setEntries(newEntries);
    await validateEntry(newEntries[index], index);
    onChange(newEntries);
  };

  const renderValueInput = (entry: IFlagMetadata, index: number) => {
    const error = errors[index]?.value;

    switch (entry.type) {
      case 'boolean':
        return (
          <Select
            value={entry.value.toString()}
            onValueChange={(value) => handleChange(index, 'value', value === 'true')}
            disabled={disabled}
          >
            <SelectTrigger
              className={`w-full ${error ? 'border-red-500' : ''}`}
              aria-invalid={!!error}
              aria-errormessage={`value-error-${index}`}
            >
              <SelectValue placeholder="Select a value" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="true">True</SelectItem>
              <SelectItem value="false">False</SelectItem>
            </SelectContent>
          </Select>
        );
      case 'number':
        return (
          <Input
            type="number"
            value={entry.value.toString()}
            onChange={(e) => handleChange(index, 'value', parseFloat(e.target.value))}
            className={error ? 'border-red-500' : ''}
            aria-invalid={!!error}
            aria-errormessage={`value-error-${index}`}
            disabled={disabled}
          />
        );
      default:
        return (
          <Input
            type="text"
            value={entry.value.toString()}
            onChange={(e) => handleChange(index, 'value', e.target.value)}
            className={error ? 'border-red-500' : ''}
            aria-invalid={!!error}
            aria-errormessage={`value-error-${index}`}
            disabled={disabled}
          />
        );
    }
  };

  return (
    <div className="space-y-4">
      {entries.map((entry, index) => {
        const keyError = errors[index]?.key;
        const typeError = errors[index]?.type;

        return (
          <div key={index} className="flex gap-4 items-start">
            <div className="flex-1">
              <Input
                type="text"
                value={entry.key}
                onChange={(e) => handleChange(index, 'key', e.target.value)}
                placeholder="Key"
                className={keyError ? 'border-red-500' : ''}
                aria-invalid={!!keyError}
                aria-errormessage={`key-error-${index}`}
                disabled={disabled}
              />
              {keyError && (
                <p className="mt-1 text-sm text-red-500" id={`key-error-${index}`}>
                  {keyError}
                </p>
              )}
            </div>

            <div className="w-32">
              <Select
                value={entry.type}
                onValueChange={(value) => handleChange(index, 'type', value)}
                disabled={disabled}
              >
                <SelectTrigger
                  className={`w-full ${typeError ? 'border-red-500' : ''}`}
                  aria-invalid={!!typeError}
                  aria-errormessage={`type-error-${index}`}
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
                <p className="mt-1 text-sm text-red-500" id={`type-error-${index}`}>
                  {typeError}
                </p>
              )}
            </div>

            <div className="flex-1">
              {renderValueInput(entry, index)}
            </div>

            <Button
              type="button"
              variant="ghost"
              size="icon"
              onClick={() => handleRemove(index)}
              disabled={disabled}
              className="mt-1"
            >
              <TrashIcon className="h-5 w-5 text-gray-500" />
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
        <PlusIcon className="h-5 w-5 mr-2" />
        Add Metadata
      </Button>
    </div>
  );
}
