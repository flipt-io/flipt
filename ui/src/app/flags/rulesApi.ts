import { TagDescription } from '@reduxjs/toolkit/query';
import { createApi } from '@reduxjs/toolkit/query/react';
import { IDistributionBase } from '~/types/Distribution';
import { IRule, IRuleBase, IRuleList } from '~/types/Rule';

import { baseQuery } from '~/utils/redux-rtk';

export const ruleTag = (arg: {
  namespaceKey: string;
  flagKey: string;
}): TagDescription<'Rule'> => {
  return { type: 'Rule', id: arg.namespaceKey + '/' + arg.flagKey };
};

export const rulesApi = createApi({
  reducerPath: 'rules',
  baseQuery,
  tagTypes: ['Rule'],
  endpoints: (builder) => ({
    // get list of rules
    listRules: builder.query<
      IRuleList,
      { namespaceKey: string; flagKey: string }
    >({
      query: ({ namespaceKey, flagKey }) =>
        `/namespaces/${namespaceKey}/flags/${flagKey}/rules`,
      providesTags: (_result, _error, arg) => [ruleTag(arg)]
    }),
    // create a new rule
    createRule: builder.mutation<
      IRule,
      {
        namespaceKey: string;
        flagKey: string;
        values: IRuleBase;
        distributions: IDistributionBase[];
      }
    >({
      queryFn: async (
        { namespaceKey, flagKey, values, distributions },
        _api,
        _extraOptions,
        baseQuery
      ) => {
        const respRule = await baseQuery({
          url: `/namespaces/${namespaceKey}/flags/${flagKey}/rules`,
          method: 'POST',
          body: values
        });
        if (respRule.error) {
          return { error: respRule.error };
        }
        const rule = respRule.data as IRule;
        const ruleId = rule.id;
        // then create the distributions
        for (let distribution of distributions) {
          const resp = await baseQuery({
            url: `/namespaces/${namespaceKey}/flags/${flagKey}/rules/${ruleId}/distributions`,
            method: 'POST',
            body: distribution
          });
          if (resp.error) {
            return { error: resp.error };
          }
        }
        return { data: rule };
      },
      invalidatesTags: (_result, _error, arg) => [ruleTag(arg)]
    }),
    // delete the rule
    deleteRule: builder.mutation<
      void,
      { namespaceKey: string; flagKey: string; ruleId: string }
    >({
      query({ namespaceKey, flagKey, ruleId }) {
        return {
          url: `/namespaces/${namespaceKey}/flags/${flagKey}/rules/${ruleId}`,
          method: 'DELETE'
        };
      },
      invalidatesTags: (_result, _error, arg) => [ruleTag(arg)]
    }),
    // update the rule
    updateRule: builder.mutation<
      IRule,
      {
        namespaceKey: string;
        flagKey: string;
        ruleId: string;
        values: IRuleBase;
      }
    >({
      query({ namespaceKey, flagKey, ruleId, values }) {
        return {
          url: `/namespaces/${namespaceKey}/flags/${flagKey}/rules/${ruleId}`,
          method: 'PUT',
          body: values
        };
      },
      invalidatesTags: (_result, _error, arg) => [ruleTag(arg)]
    }),
    // reorder the rules
    orderRules: builder.mutation<
      IRule,
      {
        namespaceKey: string;
        flagKey: string;
        ruleIds: string[];
      }
    >({
      query({ namespaceKey, flagKey, ruleIds }) {
        return {
          url: `/namespaces/${namespaceKey}/flags/${flagKey}/rules/order`,
          method: 'PUT',
          body: { ruleIds: ruleIds }
        };
      },
      async onQueryStarted(
        { namespaceKey, flagKey, ruleIds },
        { dispatch, queryFulfilled }
      ) {
        // this is manual optimistic cache update of the listRules
        // to set a desire order of rules of the listRules while server is updating the state.
        // If we don't do this we will have very strange UI state with rules in old order
        // until the result will be get from the server. It's very visible on slow connections.
        const patchResult = dispatch(
          rulesApi.util.updateQueryData(
            'listRules',
            { namespaceKey, flagKey },
            (draft: IRuleList) => {
              const rules = draft.rules;
              const resortedRules = rules.sort((a, b) => {
                const ida = ruleIds.indexOf(a.id);
                const idb = ruleIds.indexOf(b.id);
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
      invalidatesTags: (_result, _error, arg) => [ruleTag(arg)]
    }),
    // update the dustribution
    updateDistribution: builder.mutation<
      IRule,
      {
        namespaceKey: string;
        flagKey: string;
        ruleId: string;
        distributionId: string;
        values: IDistributionBase;
      }
    >({
      query({ namespaceKey, flagKey, ruleId, distributionId, values }) {
        return {
          url: `/namespaces/${namespaceKey}/flags/${flagKey}/rules/${ruleId}/distributions/${distributionId}`,
          method: 'PUT',
          body: values
        };
      },
      invalidatesTags: (_result, _error, arg) => [ruleTag(arg)]
    })
  })
});
export const {
  useListRulesQuery,
  useCreateRuleMutation,
  useDeleteRuleMutation,
  useUpdateRuleMutation,
  useOrderRulesMutation,
  useUpdateDistributionMutation
} = rulesApi;
