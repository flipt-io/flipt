import { useField } from 'formik';
import { Trash2Icon } from 'lucide-react';
import React, { KeyboardEvent, useCallback, useMemo, useState } from 'react';

import { cls } from '~/utils/helpers';

type TagsProps = {
  id: string;
  name: string;
  type?: string;
  className?: string;
  autoComplete?: boolean;
  placeholder?: string;
  forwardRef?: React.RefObject<HTMLInputElement>;
  onChange?: (e: React.ChangeEvent<HTMLInputElement>) => void;
} & React.InputHTMLAttributes<HTMLInputElement>;

type Item = string | number;

const VISIBLE_TAG_LIMIT = 50;

function FileButton({
  children,
  onSelect
}: React.PropsWithChildren<{
  onSelect: React.ChangeEventHandler<HTMLInputElement>;
}>) {
  return (
    <label className="cursor-pointer select-none rounded border border-gray-300 dark:border-gray-600 px-3 py-1 text-center text-xs text-muted-foreground focus-within:ring-2 focus-within:ring-violet-300">
      {children}
      <input
        type="file"
        accept=".txt,text/plain"
        className="sr-only"
        onChange={onSelect}
      />
    </label>
  );
}

export default function Tags(props: TagsProps) {
  const { id, type = 'text', forwardRef, placeholder } = props;
  const [field, meta] = useField(props);
  const hasError = !!(meta.touched && meta.error);
  const [inputValue, setInputValue] = useState('');
  const [inputError, setInputError] = useState('');
  const [bulkMode, setBulkMode] = useState(false);
  const [bulkValue, setBulkValue] = useState('');
  const [bulkError, setBulkError] = useState('');
  const [expanded, setExpanded] = useState(false);

  const tags: Item[] = useMemo(() => {
    try {
      const v = JSON.parse(field.value);
      if (Array.isArray(v)) {
        return v;
      }
    } catch {
      // invalid JSON — treat as empty
    }
    return [];
  }, [field.value]);

  const setFieldValue = useCallback(
    (newTags: Item[]) => {
      let v = '';
      if (newTags.length != 0) {
        v = JSON.stringify(newTags);
      }
      const e = { target: { value: v, id } };
      field.onBlur(e);
      field.onChange(e);
    },
    [field, id]
  );

  const addTag = () => {
    const raw = inputValue.trim();
    if (raw === '') return;

    let val: Item = raw;
    if (type === 'number') {
      const n = Number(raw);
      if (!Number.isFinite(n)) {
        setInputError('Must be a valid number');
        return;
      }
      val = n;
    }
    if (tags.indexOf(val) !== -1) {
      setInputError('Value already exists');
      return;
    }
    setInputError('');
    setFieldValue([...tags, val]);
    setInputValue('');
  };

  const removeTag = (tag: Item) => {
    const newTags = tags.filter((t) => t !== tag);
    setFieldValue(newTags);
  };

  const inputKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.code === 'Enter') {
      e.preventDefault();
      if (e.currentTarget.value) {
        addTag();
      }
    }
  };

  const toggleBulkMode = () => {
    setBulkValue('');
    setBulkError('');
    setBulkMode((b) => !b);
  };

  // parseBulk parses one value per line of raw into items, converting to
  // numbers for number constraints. Returns null and sets bulkError when
  // the input is empty or contains an invalid number.
  const parseBulk = (raw: string): Item[] | null => {
    const lines = raw
      .split('\n')
      .map((s) => s.trim())
      .filter(Boolean);

    if (lines.length === 0) {
      setBulkError('No values found');
      return null;
    }

    if (type === 'number') {
      const invalid = lines.filter((s) => !Number.isFinite(Number(s)));
      if (invalid.length > 0) {
        const more = invalid.length > 1 ? ` (+${invalid.length - 1} more)` : '';
        setBulkError(`Not a valid number: ${invalid[0]}${more}`);
        return null;
      }
    }

    setBulkError('');
    return lines.map((s) => (type === 'number' ? Number(s) : s));
  };

  // loadFromFile fills the bulk textarea with the picked file's contents
  // for review; one of the apply buttons then commits it.
  const loadFromFile = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    // reset so selecting the same file again re-triggers onChange
    e.target.value = '';
    if (!file) {
      return;
    }
    setBulkValue(await file.text());
    setBulkError('');
  };

  // applyBulk parses the textarea, applies update to the parsed items,
  // and returns to the regular view on success.
  const applyBulk = (update: (items: Item[]) => Item[]) => () => {
    const items = parseBulk(bulkValue);
    if (items) {
      setFieldValue(update(items));
      setBulkValue('');
      setBulkMode(false);
    }
  };

  const replaceBulk = applyBulk((items) => [...new Set(items)]);
  const addBulk = applyBulk((items) => [...new Set([...tags, ...items])]);
  const removeBulk = applyBulk((items) => {
    const remove = new Set(items);
    return tags.filter((tag) => !remove.has(tag));
  });

  const visibleTags = expanded ? tags : tags.slice(0, VISIBLE_TAG_LIMIT);
  const hiddenCount = tags.length - visibleTags.length;

  return (
    <>
      <div ref={forwardRef} id={id}>
        <div className="mb-2">
          <div className="flex items-center justify-between mb-1">
            <span className="text-xs text-gray-500 dark:text-gray-400">
              {tags.length} {tags.length === 1 ? 'value' : 'values'}
            </span>
            <button
              className="text-xs text-violet-500 hover:text-violet-600 dark:text-violet-400 dark:hover:text-violet-300"
              type="button"
              onClick={toggleBulkMode}
            >
              {bulkMode ? 'Exit bulk edit' : 'Bulk edit'}
            </button>
          </div>
          {!bulkMode && tags.length > 0 && (
            <>
              <ul className="inline-flex w-full flex-wrap gap-1">
                {visibleTags.map((tag) => (
                  <li
                    key={String(tag)}
                    className="flex flex-row items-center justify-center rounded bg-gray-200 dark:bg-gray-700 px-2 py-0.5 text-sm text-gray-900 dark:text-gray-100"
                  >
                    <span className="max-w-32 truncate" title={String(tag)}>
                      {tag}
                    </span>
                    <button
                      className="ml-1 p-1"
                      type="button"
                      onClick={() => {
                        removeTag(tag);
                      }}
                    >
                      <Trash2Icon className="h-3 w-3" aria-hidden="true" />
                    </button>
                  </li>
                ))}
              </ul>
              {hiddenCount > 0 && (
                <button
                  className="mt-1 text-xs text-violet-500 hover:text-violet-600 dark:text-violet-400 dark:hover:text-violet-300"
                  type="button"
                  onClick={() => setExpanded(true)}
                >
                  +{hiddenCount} more
                </button>
              )}
              {expanded && tags.length > VISIBLE_TAG_LIMIT && (
                <button
                  className="mt-1 ml-2 text-xs text-gray-500 hover:text-gray-600 dark:text-gray-400 dark:hover:text-gray-300"
                  type="button"
                  onClick={() => setExpanded(false)}
                >
                  Show less
                </button>
              )}
            </>
          )}
        </div>
        {bulkMode ? (
          <div className="space-y-2">
            <textarea
              className={cls(
                'block w-full rounded-md border-gray-300 bg-gray-50 text-gray-900 shadow-xs focus:border-violet-300 focus:ring-violet-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 dark:focus:border-violet-500 dark:focus:ring-violet-500 sm:text-sm',
                {
                  'border-red-400 dark:border-red-500': hasError
                }
              )}
              rows={10}
              value={bulkValue}
              placeholder="One value per line, or load from a file"
              onChange={(e) => {
                setBulkValue(e.target.value);
                setBulkError('');
              }}
            />
            {bulkError && (
              <p className="text-sm text-red-500 dark:text-red-400">
                {bulkError}
              </p>
            )}
            <div className="flex items-center justify-between">
              <div className="flex gap-2">
                <button
                  className="select-none rounded border border-violet-300 dark:border-violet-500 px-3 py-1 text-xs font-bold text-muted-foreground"
                  type="button"
                  onClick={replaceBulk}
                >
                  Replace
                </button>
                <button
                  className="select-none rounded border border-violet-300 dark:border-violet-500 px-3 py-1 text-xs font-bold text-muted-foreground"
                  type="button"
                  onClick={addBulk}
                >
                  Add
                </button>
                <button
                  className="select-none rounded border border-violet-300 dark:border-violet-500 px-3 py-1 text-xs font-bold text-muted-foreground"
                  type="button"
                  onClick={removeBulk}
                >
                  Remove
                </button>
              </div>
              <FileButton onSelect={loadFromFile}>Load from file</FileButton>
            </div>
          </div>
        ) : (
          <>
            <div className="relative flex w-full">
              <input
                className={cls(
                  'block w-full rounded-md border-gray-300 bg-gray-50 pr-20 text-gray-900 shadow-xs [appearance:textfield] focus:border-violet-300 focus:ring-violet-300 [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 dark:focus:border-violet-500 dark:focus:ring-violet-500 disabled:cursor-not-allowed disabled:border-gray-200 disabled:bg-gray-100 disabled:text-gray-500 dark:disabled:border-gray-700 dark:disabled:bg-gray-800 dark:disabled:text-gray-400 sm:text-sm',
                  {
                    'border-red-400 dark:border-red-500': hasError
                  }
                )}
                type={type}
                placeholder={placeholder}
                onKeyDown={inputKeyDown}
                value={inputValue}
                onChange={(e) => {
                  setInputValue(e.currentTarget.value);
                  setInputError('');
                }}
              />
              <button
                className="z-1 border-1 absolute! right-1 top-1 select-none rounded border border-violet-300 dark:border-violet-500 px-4 py-1.5 text-center align-middle text-xs font-bold text-muted-foreground"
                type="button"
                onClick={addTag}
              >
                Add
              </button>
            </div>
            {inputError && (
              <p className="mt-1 text-sm text-red-500 dark:text-red-400">
                {inputError}
              </p>
            )}
          </>
        )}
      </div>
      {hasError && meta.error?.length && meta.error.length > 0 ? (
        <div className="mt-1 text-sm text-red-500 dark:text-red-400">
          {meta.error}
        </div>
      ) : null}
    </>
  );
}
