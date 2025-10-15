import { useField } from 'formik';
import { Trash2Icon } from 'lucide-react';
import React, { KeyboardEvent, useState } from 'react';

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

export default function Tags(props: TagsProps) {
  const { id, type = 'text', forwardRef, placeholder } = props;
  const [field, meta] = useField(props);
  const hasError = !!(meta.touched && meta.error);
  const [inputValue, setInputValue] = useState('');
  let tags: Item[] = [];
  try {
    const v = JSON.parse(field.value);
    if (Array.isArray(v)) {
      tags = v;
    }
  } catch (e) {
    // todo - what could be done?
  }

  const setFieldValue = (newTags: Item[]) => {
    let v = '';
    if (newTags.length != 0) {
      v = JSON.stringify(newTags);
    }
    const e = { target: { value: v, id } };
    field.onBlur(e);
    field.onChange(e);
  };

  const addTag = () => {
    let val: Item = inputValue.trim();
    if (type == 'number') {
      val = Number(val);
    }
    if (!val || tags.indexOf(val) != -1) {
      return;
    }
    setFieldValue([...tags, val]);
    setInputValue('');
  };

  const removeTag = (i: number) => {
    const newTags = [...tags];
    newTags.splice(i, 1);
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

  return (
    <>
      <div ref={forwardRef} id={id}>
        {tags.length > 0 && (
          <ul className="mb-2 inline-flex w-full flex-wrap gap-1">
            {tags.map((tag, i) => (
              <li
                key={tag}
                className="flex flex-row items-center justify-center rounded bg-gray-200 dark:bg-gray-700 px-2 py-0.5 text-sm text-gray-900 dark:text-gray-100"
              >
                <span className="max-w-32 truncate" title={String(tag)}>
                  {tag}
                </span>
                <button
                  className="ml-1 p-1"
                  type="button"
                  onClick={() => {
                    removeTag(i);
                  }}
                >
                  <Trash2Icon className="h-3 w-3" aria-hidden="true" />
                </button>
              </li>
            ))}
          </ul>
        )}
        <div className="relative flex w-full">
          <input
            className={cls(
              'block w-full rounded-md border-gray-300 bg-gray-50 text-gray-900 shadow-xs focus:border-violet-300 focus:ring-violet-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 dark:focus:border-violet-500 dark:focus:ring-violet-500 disabled:cursor-not-allowed disabled:border-gray-200 disabled:bg-gray-100 disabled:text-gray-500 dark:disabled:border-gray-700 dark:disabled:bg-gray-800 dark:disabled:text-gray-400 sm:text-sm',
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
      </div>
      {hasError && meta.error?.length && meta.error.length > 0 ? (
        <div className="mt-1 text-sm text-red-500 dark:text-red-400">
          {meta.error}
        </div>
      ) : null}
    </>
  );
}
