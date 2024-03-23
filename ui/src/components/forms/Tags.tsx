import { useField } from 'formik';
import { cls } from '~/utils/helpers';
import { TrashIcon } from '@heroicons/react/24/outline';
import React, { KeyboardEvent, useState } from 'react';

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
    if (e.code === 'Enter' && e.currentTarget.value) {
      e.preventDefault();
      addTag();
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
                className="text-gray-900 bg-gray-200 flex flex-row items-center justify-center rounded px-2 py-0.5 text-sm"
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
                  <TrashIcon className="h-3 w-3" aria-hidden="true" />
                </button>
              </li>
            ))}
          </ul>
        )}
        <div className="relative flex w-full">
          <input
            className={cls(
              'text-gray-900 bg-gray-50 border-gray-300 block w-full rounded-md shadow-sm focus:border-violet-300 disabled:text-gray-500 disabled:bg-gray-100 disabled:border-gray-200 focus:ring-violet-300 disabled:cursor-not-allowed sm:text-sm',
              {
                'border-red-400': hasError
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
            className="z-1 border-1 bg-white text-gray-500 border-violet-300 !absolute right-1 top-1 select-none rounded border px-4 py-1.5 text-center align-middle text-xs font-bold"
            type="button"
            onClick={addTag}
          >
            Add
          </button>
        </div>
      </div>
      {hasError && meta.error?.length && meta.error.length > 0 ? (
        <div className="text-red-500 mt-1 text-sm">{meta.error}</div>
      ) : null}
    </>
  );
}
