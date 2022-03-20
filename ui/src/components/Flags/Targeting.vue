<template>
  <section class="section">
    <div class="container">
      <nav class="breadcrumb" aria-label="breadcrumbs">
        <ul>
          <li><RouterLink :to="{ name: 'flags' }">Flags</RouterLink></li>
          <li class="is-active">
            <a href="#" aria-current="page">{{ flag.key }}</a>
          </li>
        </ul>
      </nav>
      <div class="tabs is-boxed">
        <ul>
          <li>
            <RouterLink
              :to="{ name: 'flag', params: { key: $route.params.key } }"
              >Details</RouterLink
            >
          </li>
          <li class="is-active">
            <RouterLink
              :to="{ name: 'targeting', params: { key: $route.params.key } }"
              >Targeting</RouterLink
            >
          </li>
        </ul>
      </div>
      <h5 class="title is-5">Rules</h5>
      <p class="subtitle is-7">
        Evaluated in order from top to bottom. Reorder rules by dragging and
        dropping them in place.
      </p>
      <draggable v-model="rules" @update="reordered = true">
        <Rule
          v-for="(rule, index) in rules"
          :key="rule.id"
          v-bind="rule"
          :index="index"
          @deleteRule="deleteRule(index)"
          @editRule="editRule(index)"
        />
      </draggable>
      <br />
      <div class="field is-grouped">
        <div class="control">
          <button
            class="button is-primary"
            @click.prevent="dialogAddRuleVisible = true"
          >
            New Rule
          </button>
          <button
            class="button"
            :disabled="!reordered"
            @click.prevent="reorderRules"
          >
            Reorder Rules
          </button>
        </div>
      </div>
      <hr />
      <DebugConsole :flag="flag" />
    </div>

    <div
      id="addRuleDialog"
      class="modal"
      tabindex="0"
      :class="{ 'is-active': dialogAddRuleVisible }"
      @keyup.esc="cancelAddRule"
    >
      <div class="modal-background" @click.prevent="cancelAddRule" />
      <div class="modal-content">
        <div class="container">
          <div class="box">
            <form>
              <div class="field is-horizontal">
                <div class="field-label is-normal">
                  <strong>IF</strong> Matches Segment:
                </div>
                <div class="field-body">
                  <div class="field">
                    <b-autocomplete
                      v-model="newRule.segmentName"
                      :data="autocompleteSegmentData"
                      field="name"
                      :open-on-focus="true"
                      placeholder="e.g. All Users"
                      @select="(option) => selectSegment(option)"
                    >
                      <template slot="empty"
                        >No segments found. Create a
                        <RouterLink :to="{ name: 'new-segment' }"
                          >New Segment</RouterLink
                        >.</template
                      >
                    </b-autocomplete>
                  </div>
                </div>
              </div>
              <template v-if="hasVariants">
                <div class="field is-horizontal">
                  <div class="field-label is-normal">
                    <strong>THEN</strong> Serve Variant(s):
                  </div>
                  <div class="field-body">
                    <div class="field">
                      <div class="select">
                        <select
                          v-model="selectedVariant"
                          :disabled="!newRule.segmentKey"
                          @change="ruleTypeChanged"
                        >
                          <option value="">Choose Value</option>
                          <option disabled>──────────</option>
                          <option
                            v-for="variant in flag.variants"
                            :key="variant.id"
                            :value="variant.id"
                          >
                            {{ variant.key }}
                          </option>
                          <option disabled>──────────</option>
                          <option value="rollout">A Percentage Rollout</option>
                        </select>
                      </div>
                    </div>
                  </div>
                </div>
                <template v-if="newRule.distributions.length > 1">
                  <hr />
                  <div
                    v-for="(variant, index) in flag.variants"
                    :key="index"
                    class="field is-horizontal"
                  >
                    <div class="field-label">
                      <span class="tag is-small">{{ variant.key }}</span>
                    </div>
                    <div class="field-body">
                      <div class="field">
                        <b-input
                          v-model="newRule.distributions[index].rollout"
                          placeholder="Percentage"
                          type="number"
                          icon-pack="fas"
                          icon="percent"
                          size="is-small"
                          min="0"
                          max="100"
                        />
                      </div>
                    </div>
                  </div>
                </template>
              </template>
              <div class="field is-grouped">
                <div class="control">
                  <button
                    class="button is-primary"
                    :disabled="!canAddRule"
                    @click.prevent="addRule"
                  >
                    Add Rule
                  </button>
                  <button class="button is-text" @click.prevent="cancelAddRule">
                    Cancel
                  </button>
                </div>
              </div>
            </form>
          </div>
          <button
            class="modal-close is-large"
            aria-label="close"
            @click.prevent="cancelAddRule"
          />
        </div>
      </div>
    </div>

    <div
      id="editRule"
      class="modal"
      tabindex="0"
      :class="{ 'is-active': dialogEditRuleVisible }"
      @keyup.esc="cancelEditRule"
    >
      <div class="modal-background" @click.prevent="cancelEditRule" />
      <div class="modal-content">
        <div class="container">
          <div class="box">
            <form>
              <div class="field is-horizontal">
                <div class="field-label is-normal">Segment:</div>
                <div class="field-body">
                  <div class="field">
                    <span class="tag is-medium">
                      {{ selectedRule.segmentName }}
                    </span>
                  </div>
                </div>
              </div>
              <div class="field is-horizontal">
                <div class="field-label is-normal">Then serve:</div>
                <div class="field-body">
                  <div class="field">
                    <div class="select">
                      <select disabled>
                        <option selected>A Percentage Rollout</option>
                      </select>
                    </div>
                  </div>
                </div>
              </div>
              <hr />
              <div
                v-for="(distribution, index) in selectedRule.distributions"
                :key="index"
                class="field is-horizontal"
              >
                <div class="field-label">
                  <span class="tag is-small">
                    {{ distribution.variantKey }}
                  </span>
                </div>
                <div class="field-body">
                  <div class="field">
                    <b-input
                      v-model="selectedRule.distributions[index].rollout"
                      placeholder="Percentage"
                      type="number"
                      icon-pack="fas"
                      icon="percent"
                      size="is-small"
                      min="0"
                      max="100"
                    />
                  </div>
                </div>
              </div>
              <div class="field is-grouped">
                <div class="control">
                  <button class="button is-primary" @click.prevent="updateRule">
                    Update Rule
                  </button>
                  <button
                    class="button is-text"
                    @click.prevent="cancelEditRule"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            </form>
          </div>
          <button
            class="modal-close is-large"
            aria-label="close"
            @click.prevent="cancelEditRule"
          />
        </div>
      </div>
    </div>
  </section>
</template>

<script>
import {
  capitalize,
  clone,
  cloneDeep,
  find,
  forEach,
  isEmpty,
  map,
} from "lodash";

import draggable from "vuedraggable";

import { Api } from "@/services/api";
import notify from "@/mixins/notify";
import targeting from "@/mixins/targeting";
import utils from "@/mixins/utils";
import Rule from "./Rule";
import DebugConsole from "./DebugConsole";

const DEFAULT_RULE = {
  segmentKey: "",
  segmentName: "",
  distributions: [],
  rank: 0,
};

export default {
  name: "FlagTargeting",
  components: {
    draggable,
    Rule,
    DebugConsole,
  },
  mixins: [notify, targeting, utils],
  beforeRouteLeave(to, from, next) {
    if (this.reordered === false) {
      next();
      return;
    }
    if (
      confirm(
        "Are you sure want to leave? Rules have been reordered but not saved."
      )
    ) {
      next();
    } else {
      next(false);
    }
  },
  data() {
    return {
      dialogAddRuleVisible: false,
      dialogEditRuleVisible: false,
      flag: {
        variants: [],
      },
      rules: [],
      reordered: false,
      segments: [],
      newRule: clone(DEFAULT_RULE),
      selectedRule: clone(DEFAULT_RULE),
      selectedVariant: "",
    };
  },
  computed: {
    autocompleteSegmentData() {
      return this.segments.filter((option) => {
        return (
          option.name
            .toString()
            .toLowerCase()
            .indexOf(this.newRule.segmentName.toLowerCase()) >= 0
        );
      });
    },
    hasVariants() {
      return this.flag.variants && this.flag.variants.length > 0;
    },
    canAddRule() {
      return (
        this.isPresent(this.newRule.segmentKey) &&
        !isEmpty(this.newRule.distributions)
      );
    },
  },
  mounted() {
    this.fetchData();
  },
  methods: {
    fetchData() {
      Api.get("/segments")
        .then((response) => {
          this.segments = response.data.segments;

          let key = this.$route.params.key;
          Api.get("/flags/" + key)
            .then((response) => {
              this.flag = response.data;

              Api.get("/flags/" + key + "/rules").then((response) => {
                this.rules = map(response.data.rules, (r) => {
                  return this.processRule(r);
                });
              });
            })
            .catch(() => {
              this.notifyError("Error loading flag.");
              this.$router.push("/flags");
            });
        })
        .catch((error) => {
          this.notifyError("Error loading data.");
          console.error(error);
        });
    },
    selectSegment(option) {
      if (!option) {
        return;
      }
      this.newRule.segmentKey = option.key;
    },
    addRule() {
      if (!this.validRollout(this.newRule.distributions)) {
        this.notifyError("Total distribution percentage cannot exceed 100%.");
        return;
      }

      this.newRule.rank = this.rules.length + 1;

      // create the rule
      Api.post("/flags/" + this.flag.key + "/rules", this.newRule)
        .then((response) => {
          let rule = response.data;
          let segment = find(this.segments, { key: rule.segmentKey });
          rule.segmentName = segment.name;
          rule.distributions = [];

          // create the distributions
          forEach(this.newRule.distributions, (d) => {
            Api.post(
              "/flags/" +
                this.flag.key +
                "/rules/" +
                rule.id +
                "/distributions",
              d
            ).then((response) => {
              let distribution = response.data;
              let variant = find(this.flag.variants, {
                id: distribution.variantId,
              });
              distribution.variantKey = variant.key;
              rule.distributions.push(distribution);
            });
          });

          this.rules.push(rule);
          this.newRule = clone(DEFAULT_RULE);
          this.notifySuccess("Rule added!");
          this.dialogAddRuleVisible = false;
          this.selectedVariant = "";
        })
        .catch((error) => {
          if (error.response && error.response.data) {
            this.notifyError(capitalize(error.response.data.message));
          } else {
            this.notifyError("Error creating rule.");
            console.error(error);
          }
        });
    },
    cancelAddRule() {
      this.dialogAddRuleVisible = false;
      this.selectedVariant = "";
      this.newRule = clone(DEFAULT_RULE);
    },
    updateRule() {
      if (!this.validRollout(this.selectedRule.distributions)) {
        this.notifyError("Total distribution percentage cannot exceed 100%.");
        return;
      }

      try {
        forEach(this.selectedRule.distributions, (d) => {
          Api.put(
            "/flags/" +
              this.flag.key +
              "/rules/" +
              this.selectedRule.id +
              "/distributions/" +
              d.id,
            d
          );
        });

        let index = this.rules.findIndex((r) => r.id === this.selectedRule.id);
        this.rules[index] = clone(this.selectedRule);
        this.selectedRule = clone(DEFAULT_RULE);
        this.notifySuccess("Rule updated!");
        this.dialogEditRuleVisible = false;
      } catch (error) {
        if (error.response && error.response.data) {
          this.notifyError(capitalize(error.response.data.message));
        } else {
          this.notifyError("Error updating rule.");
          console.error(error);
        }
      }
    },
    deleteRule(index) {
      if (!confirm(`Are you sure you want to delete this rule?`)) {
        return;
      }

      let rule = this.rules[index];

      Api.delete("/flags/" + this.flag.key + "/rules/" + rule.id)
        .then(() => {
          this.rules.splice(index, 1);
        })
        .catch((error) => {
          this.notifyError("Error deleting rule.");
          console.error(error);
        });
    },
    editRule(index) {
      this.dialogEditRuleVisible = true;
      this.selectedRule = cloneDeep(this.rules[index]);
    },
    cancelEditRule() {
      this.dialogEditRuleVisible = false;
      this.selectedRule = clone(DEFAULT_RULE);
    },
    reorderRules() {
      this.reordered = false;

      Api.put("/flags/" + this.flag.key + "/rules/order", {
        ruleIds: map(this.rules, "id"),
      })
        .then(() => {
          this.notifySuccess("Rules reordered!");
        })
        .catch((error) => {
          if (error.response && error.response.data) {
            this.notifyError(capitalize(error.response.data.message));
          } else {
            this.notifyError("Error ordering rules.");
            console.error(error);
          }
        });
    },
    processRule(rule) {
      let segment = find(this.segments, { key: rule.segmentKey });

      rule.segmentName = segment.name;
      rule.distributions = map(rule.distributions, (d) => {
        let variant = find(this.flag.variants, { id: d.variantId });
        d.variantKey = variant.key;
        return d;
      });
      return rule;
    },
    ruleTypeChanged(event) {
      let val = event.target.value;
      if (val === "") {
        this.newRule.distributions = [];
      } else if (val === "rollout") {
        let n = this.flag.variants.length;
        let percentages = this.computePercentages(n);
        let distributions = [];

        for (let i = 0; i < n; i++) {
          let v = this.flag.variants[i];

          distributions.push({
            variantId: v.id,
            variantKey: v.key,
            rollout: percentages[i],
          });
        }

        this.newRule.distributions = distributions;
      } else {
        let variant = find(this.flag.variants, { id: val });
        this.newRule.distributions = [
          {
            variantId: variant.id,
            variantKey: variant.key,
            rollout: 100,
          },
        ];
      }
    },
  },
};
</script>
