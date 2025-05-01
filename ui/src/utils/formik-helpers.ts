import { FormikContextType } from 'formik';

import { FilterableSegment } from '~/types/Segment';

interface WithIdAndSegments {
  id?: string;
  segments?: string[];
  rank: number;
}

// Define a type for the shape of formik values used in collections
interface FormikValuesWithCollection<T> {
  [key: string]: T[] | any;
}

/**
 * Updates a collection item by ID or rank with new segment information.
 * This helps manage segment operations consistently across different components.
 *
 * @param formik The Formik context to use for updates
 * @param item The current item (rule/rollout) being updated
 * @param currentSegments The array of segment keys to set
 * @param collectionName The name of the collection in formik values (e.g. 'rules', 'rollouts')
 * @param getUpdatedItem A function that returns the updated item with the new segments
 */
export function updateItemSegments<
  T extends WithIdAndSegments,
  V extends FormikValuesWithCollection<T>
>(
  formik: FormikContextType<V>,
  item: T,
  currentSegments: string[],
  collectionName: string,
  getUpdatedItem: (original: T, segments: string[]) => T
): void {
  // Use the correct item path to ensure we're updating the existing item
  const fieldPrefix = `${collectionName}.[${item.rank}].`;

  if (item.id) {
    // First check if the item already exists in the formik values
    const collection = (formik.values[collectionName] || []) as T[];
    const itemIndex = collection.findIndex((r: T) => r.id === item.id);

    if (itemIndex >= 0) {
      // Update the existing item
      const updatedCollection = [...collection];
      updatedCollection[itemIndex] = getUpdatedItem(
        updatedCollection[itemIndex],
        currentSegments
      );

      // Update all items at once to prevent duplicates
      formik.setFieldValue(collectionName, updatedCollection);
    } else {
      // Fallback to updating by rank if ID is not found
      formik.setFieldValue(`${fieldPrefix}segments`, currentSegments);
    }
  } else {
    // Fallback to the original method if no ID
    formik.setFieldValue(`${fieldPrefix}segments`, currentSegments);
  }
}

/**
 * Creates handlers for segment operations (add, remove, replace) to be used with SegmentsPicker
 *
 * @param formik The Formik context
 * @param item The current item (rule/rollout) with segments
 * @param collectionName The name of the collection in formik values (e.g. 'rules', 'rollouts')
 * @param getUpdatedItem Function to create an updated item with new segments
 * @param initFn Optional initialization function to call before adding segments
 */
export function createSegmentHandlers<
  T extends WithIdAndSegments,
  V extends FormikValuesWithCollection<T>
>(
  formik: FormikContextType<V>,
  item: T,
  collectionName: string,
  getUpdatedItem: (original: T, segments: string[]) => T,
  initFn?: () => void
) {
  const handleSegmentAdd = (segment: FilterableSegment) => {
    // Run the initialization function if provided
    if (initFn) initFn();

    // Get the current segments array
    const currentSegments = [...(item.segments || [])];

    // Add the new segment key to the array
    currentSegments.push(segment.key);

    updateItemSegments(
      formik,
      item,
      currentSegments,
      collectionName,
      getUpdatedItem
    );
  };

  const handleSegmentRemove = (index: number) => {
    // Get the current segments array
    const currentSegments = [...(item.segments || [])];

    // Remove the segment at the specified index
    currentSegments.splice(index, 1);

    updateItemSegments(
      formik,
      item,
      currentSegments,
      collectionName,
      getUpdatedItem
    );
  };

  const handleSegmentReplace = (index: number, segment: FilterableSegment) => {
    // Get the current segments array
    const currentSegments = [...(item.segments || [])];

    // Replace the segment at the specified index with the new segment key
    currentSegments[index] = segment.key;

    updateItemSegments(
      formik,
      item,
      currentSegments,
      collectionName,
      getUpdatedItem
    );
  };

  return {
    handleSegmentAdd,
    handleSegmentRemove,
    handleSegmentReplace
  };
}
