import { TagDescription } from '@reduxjs/toolkit/query';
import { createApi } from '@reduxjs/toolkit/query/react';
import { IRollout, IRolloutBase, IRolloutList } from '~/types/Rollout';

import { baseQuery } from '~/utils/redux-rtk';

export const rolloutTag = (arg: {
  namespaceKey: string;
  flagKey: string;
}): TagDescription<'Rollout'> => {
  return { type: 'Rollout', id: arg.namespaceKey + '/' + arg.flagKey };
};

export const rolloutsApi = createApi({
  reducerPath: 'rollouts',
  baseQuery,
  tagTypes: ['Rollout'],
  endpoints: (builder) => ({
    // get list of rollouts
    listRollouts: builder.query<
      IRolloutList,
      { namespaceKey: string; flagKey: string }
    >({
      query: ({ namespaceKey, flagKey }) =>
        `/namespaces/${namespaceKey}/flags/${flagKey}/rollouts`,
      providesTags: (_result, _error, arg) => [rolloutTag(arg)]
    }),
    // delete the rollout
    deleteRollout: builder.mutation<
      void,
      { namespaceKey: string; flagKey: string; rolloutId: string }
    >({
      query({ namespaceKey, flagKey, rolloutId }) {
        return {
          url: `/namespaces/${namespaceKey}/flags/${flagKey}/rollouts/${rolloutId}`,
          method: 'DELETE'
        };
      },
      invalidatesTags: (_result, _error, arg) => [rolloutTag(arg)]
    }),
    // create the rollout
    createRollout: builder.mutation<
      void,
      {
        namespaceKey: string;
        flagKey: string;
        values: IRolloutBase;
      }
    >({
      query({ namespaceKey, flagKey, values }) {
        return {
          url: `/namespaces/${namespaceKey}/flags/${flagKey}/rollouts`,
          method: 'POST',
          body: values
        };
      },
      invalidatesTags: (_result, _error, arg) => [rolloutTag(arg)]
    }),
    // update the rollout
    updateRollout: builder.mutation<
      void,
      {
        namespaceKey: string;
        flagKey: string;
        rolloutId: string;
        values: IRolloutBase;
      }
    >({
      query({ namespaceKey, flagKey, rolloutId, values }) {
        return {
          url: `/namespaces/${namespaceKey}/flags/${flagKey}/rollouts/${rolloutId}`,
          method: 'PUT',
          body: values
        };
      },
      invalidatesTags: (_result, _error, arg) => [rolloutTag(arg)]
    }),
    // reorder the rollouts
    orderRollouts: builder.mutation<
      IRollout,
      {
        namespaceKey: string;
        flagKey: string;
        rolloutIds: string[];
      }
    >({
      query({ namespaceKey, flagKey, rolloutIds }) {
        return {
          url: `/namespaces/${namespaceKey}/flags/${flagKey}/rollouts/order`,
          method: 'PUT',
          body: { rolloutIds: rolloutIds }
        };
      },
      async onQueryStarted(
        { namespaceKey, flagKey, rolloutIds },
        { dispatch, queryFulfilled }
      ) {
        // this is manual optimistic cache update of the listRollouts
        // to set a desire order of rules of the listRollouts while server is updating the state.
        // If we don't do this we will have very strange UI state with rules in old order
        // until the result will be get from the server. It's very visible on slow connections.
        const patchResult = dispatch(
          rolloutsApi.util.updateQueryData(
            'listRollouts',
            { namespaceKey, flagKey },
            (draft: IRolloutList) => {
              const rules = draft.rules;
              const resortedRules = rules.sort((a, b) => {
                const ida = rolloutIds.indexOf(a.id);
                const idb = rolloutIds.indexOf(b.id);
                if (ida < idb) {
                  return -1;
                } else if (ida > idb) {
                  return 1;
                }
                return 0;
              });
              return Object.assign(draft, { rules: resortedRules });
            }
          )
        );
        try {
          await queryFulfilled;
        } catch {
          patchResult.undo();
        }
      },
      invalidatesTags: (_result, _error, arg) => [rolloutTag(arg)]
    })
  })
});

export const {
  useListRolloutsQuery,
  useCreateRolloutMutation,
  useUpdateRolloutMutation,
  useDeleteRolloutMutation,
  useOrderRolloutsMutation
} = rolloutsApi;
