<template>
  <div>
    <section class="section">
      <div class="container">
        <nav class="breadcrumb" aria-label="breadcrumbs">
          <ul>
            <li>
              <RouterLink :to="{ name: 'segments' }">Segments</RouterLink>
            </li>
            <li class="is-active">
              <a href="#" aria-current="page">{{ segment.key }}</a>
            </li>
          </ul>
        </nav>
        <form>
          <b-field label="Key">
            <b-input v-model="segment.key" disabled />
          </b-field>
          <b-field label="Name">
            <b-input
              v-model="segment.name"
              placeholder="Segment name"
              required
            />
          </b-field>
          <b-field label="Description (optional)">
            <b-input
              v-model="segment.description"
              placeholder="Segment description"
            />
          </b-field>
          <b-field label="Match Type" :message="matchTypeText">
            <div class="block">
              <b-radio
                v-model="segment.matchType"
                native-value="ALL_MATCH_TYPE"
              >
                Match All
              </b-radio>
              <b-radio
                v-model="segment.matchType"
                native-value="ANY_MATCH_TYPE"
              >
                Match Any
              </b-radio>
            </div>
          </b-field>
          <hr />
          <div class="level">
            <div class="level-left">
              <div class="level-item">
                <div class="field is-grouped">
                  <div class="control">
                    <button
                      class="button is-primary"
                      :disabled="!canUpdateSegment"
                      @click.prevent="updateSegment()"
                    >
                      Save
                    </button>
                  </div>
                  <div class="control">
                    <RouterLink
                      class="button is-text"
                      :to="{ name: 'segments' }"
                      >Cancel</RouterLink
                    >
                  </div>
                </div>
              </div>
            </div>
            <div class="level-right">
              <div class="level-item">
                <div class="field is-grouped">
                  <div class="control">
                    <button
                      class="button is-danger"
                      @click.prevent="dialogDeleteSegmentVisible = true"
                    >
                      Delete
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </form>
      </div>
    </section>
    <section class="section">
      <div class="container">
        <h5 class="title is-5">Constraints</h5>
        <p class="subtitle is-7">Determine if an entity matches your segment</p>
        <b-table :data="segment.constraints">
          <b-table-column
            v-slot="props"
            field="property"
            label="Property"
            sortable
          >
            {{ props.row.property }}
          </b-table-column>
          <b-table-column v-slot="props" field="type" label="Type" sortable>
            {{ comparisons[props.row.type] }}
          </b-table-column>
          <b-table-column
            v-slot="props"
            field="operator"
            label="Operator"
            centered
          >
            {{ allOperators[props.row.operator] }}
          </b-table-column>
          <b-table-column v-slot="props" field="value" label="Value">
            {{ props.row.value }}
          </b-table-column>
          <b-table-column v-slot="props" field="" label="" width="160" centered>
            <a
              class="button is-white"
              @click.prevent="editConstraint(props.index)"
            >
              <span class="icon is-small">
                <i class="fas fa-pencil-alt" title="Edit" />
              </span>
            </a>
            <a
              class="button is-white"
              @click.prevent="duplicateConstraint(props.index)"
            >
              <span class="icon is-small">
                <i class="far fa-clone" title="Duplicate" />
              </span>
            </a>
            <a
              class="button is-white"
              @click.prevent="deleteConstraint(props.index)"
            >
              <span class="icon is-small">
                <i class="fas fa-times" title="Delete" />
              </span>
            </a>
          </b-table-column>
        </b-table>
        <br />
        <div class="field">
          <div class="control">
            <button
              class="button is-primary"
              @click.prevent="dialogAddConstraintVisible = true"
            >
              New Constraint
            </button>
          </div>
        </div>
      </div>
    </section>

    <div
      id="addConstraintDialog"
      class="modal"
      tabindex="0"
      :class="{ 'is-active': dialogAddConstraintVisible }"
      @keyup.esc="cancelAddConstraint"
    >
      <div class="modal-background" @click.prevent="cancelAddConstraint" />
      <div class="modal-content">
        <div class="container">
          <div class="box">
            <form>
              <b-field label="Property">
                <b-input
                  v-model="newConstraint.property"
                  placeholder="Property"
                  required
                />
              </b-field>
              <b-field label="Comparison Type">
                <BSelect
                  v-model="newConstraint.type"
                  placeholder="Select a type"
                >
                  <option
                    v-for="(value, key, index) in comparisons"
                    :key="index"
                    :value="key"
                  >
                    {{ value }}
                  </option>
                </BSelect>
              </b-field>
              <b-field label="Operator">
                <BSelect
                  v-model="newConstraint.operator"
                  placeholder="Select an operator"
                  :disabled="!newConstraint.type"
                >
                  <option
                    v-for="(value, key, index) in operators(newConstraint.type)"
                    :key="index"
                    :value="key"
                  >
                    {{ value }}
                  </option>
                </BSelect>
              </b-field>
              <b-field v-show="hasValue(newConstraint)" label="Value">
                <b-input
                  v-model="newConstraint.value"
                  placeholder="Value"
                  required
                />
              </b-field>
              <div class="field is-grouped">
                <div class="control">
                  <button
                    class="button is-primary"
                    :disabled="!canAddConstraint"
                    @click.prevent="addConstraint"
                  >
                    Add Constraint
                  </button>
                  <button
                    class="button is-text"
                    @click.prevent="cancelAddConstraint"
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
            @click.prevent="cancelAddConstraint"
          />
        </div>
      </div>
    </div>

    <div
      id="editConstraintDialog"
      class="modal"
      tabindex="0"
      :class="{ 'is-active': dialogEditConstraintVisible }"
      @keyup.esc="cancelEditConstraint"
    >
      <div class="modal-background" @click.prevent="cancelEditConstraint" />
      <div class="modal-content">
        <div class="container">
          <div class="box">
            <form>
              <b-field label="Property">
                <b-input
                  v-model="selectedConstraint.property"
                  placeholder="Property"
                  required
                />
              </b-field>
              <b-field label="Comparison Type">
                <BSelect
                  v-model="selectedConstraint.type"
                  placeholder="Select a type"
                >
                  <option
                    v-for="(value, key, index) in comparisons"
                    :key="index"
                    :value="key"
                  >
                    {{ value }}
                  </option>
                </BSelect>
              </b-field>
              <b-field label="Operator">
                <BSelect
                  v-model="selectedConstraint.operator"
                  placeholder="Select an operator"
                >
                  <option
                    v-for="(value, key, index) in operators(
                      selectedConstraint.type
                    )"
                    :key="index"
                    :value="key"
                  >
                    {{ value }}
                  </option>
                </BSelect>
              </b-field>
              <b-field v-show="hasValue(selectedConstraint)" label="Value">
                <b-input
                  v-model="selectedConstraint.value"
                  placeholder="Value"
                  required
                />
              </b-field>
              <div class="field is-grouped">
                <div class="control">
                  <button
                    class="button is-primary"
                    :disabled="!canUpdateConstraint"
                    @click.prevent="updateConstraint"
                  >
                    Update Constraint
                  </button>
                  <button
                    class="button is-text"
                    @click.prevent="cancelEditConstraint"
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
            @click.prevent="cancelEditConstraint"
          />
        </div>
      </div>
    </div>

    <div
      id="deleteSegmentDialog"
      class="modal"
      tabindex="0"
      :class="{ 'is-active': dialogDeleteSegmentVisible }"
      @keyup.esc="dialogDeleteSegmentVisible = false"
    >
      <div
        class="modal-background"
        @click="dialogDeleteSegmentVisible = false"
      />
      <div class="modal-content">
        <div class="container">
          <div class="box">
            <p class="has-text-centered">
              Are you sure you want to delete this segment?
            </p>
            <br />
            <div class="control has-text-centered">
              <button class="button is-danger" @click.prevent="deleteSegment">
                Confirm
              </button>
              <button
                class="button is-text"
                @click="dialogDeleteSegmentVisible = false"
              >
                Cancel
              </button>
            </div>
          </div>
          <button
            class="modal-close is-large"
            aria-label="close"
            @click="dialogDeleteSegmentVisible = false"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { capitalize, clone, cloneDeep, merge } from "lodash";

import { Api } from "@/services/api";
import notify from "@/mixins/notify";
import utils from "@/mixins/utils";

const STRING_OPERATORS = {
  eq: "==",
  neq: "!=",
  empty: "IS EMPTY",
  notempty: "IS NOT EMPTY",
  prefix: "HAS PREFIX",
  suffix: "HAS SUFFIX",
};

const NUMBER_OPERATORS = {
  eq: "==",
  neq: "!=",
  lt: "<",
  lte: "<=",
  gt: ">",
  gte: ">=",
  present: "IS PRESENT",
  notpresent: "IS NOT PRESENT",
};

const BOOLEAN_OPERATORS = {
  true: "IS TRUE",
  false: "IS FALSE",
  present: "IS PRESENT",
  notpresent: "IS NOT PRESENT",
};

const COMPARISONS = {
  STRING_COMPARISON_TYPE: "string",
  NUMBER_COMPARISON_TYPE: "number",
  BOOLEAN_COMPARISON_TYPE: "boolean",
};

const DEFAULT_CONSTRAINT = {
  type: "STRING_COMPARISON_TYPE",
  property: "",
  operator: "eq",
  value: "",
};

export default {
  name: "Segment",
  mixins: [notify, utils],
  data() {
    return {
      dialogDeleteSegmentVisible: false,
      dialogAddConstraintVisible: false,
      dialogEditConstraintVisible: false,
      segment: {
        matchType: "ALL_MATCH_TYPE",
        constraints: [],
      },
      newConstraint: clone(DEFAULT_CONSTRAINT),
      selectedConstraint: clone(DEFAULT_CONSTRAINT),
      comparisons: COMPARISONS,
      allOperators: merge(
        {},
        STRING_OPERATORS,
        NUMBER_OPERATORS,
        BOOLEAN_OPERATORS
      ),
    };
  },
  computed: {
    canUpdateSegment() {
      return (
        this.isPresent(this.segment.key) && this.isPresent(this.segment.name)
      );
    },
    canAddConstraint() {
      let valid =
        this.isPresent(this.newConstraint.property) &&
        this.isPresent(this.newConstraint.type) &&
        this.isPresent(this.newConstraint.operator);

      if (this.hasValue(this.newConstraint)) {
        return valid && this.isPresent(this.newConstraint.value);
      }

      return valid;
    },
    canUpdateConstraint() {
      let valid =
        this.isPresent(this.selectedConstraint.property) &&
        this.isPresent(this.selectedConstraint.type) &&
        this.isPresent(this.selectedConstraint.operator);

      if (this.hasValue(this.selectedConstraint)) {
        return valid && this.isPresent(this.selectedConstraint.value);
      }

      return valid;
    },
    matchTypeText() {
      if (this.segment.matchType === "ALL_MATCH_TYPE") {
        return "All constraints must match.";
      } else {
        return "At least one constraint must match.";
      }
    },
  },
  mounted() {
    this.getSegment();
  },
  methods: {
    formatKey() {
      this.segment.key = this.segment.key.toLowerCase().split(" ").join("-");
    },
    operators(type) {
      switch (type) {
        case "STRING_COMPARISON_TYPE":
          return STRING_OPERATORS;

        case "NUMBER_COMPARISON_TYPE":
          return NUMBER_OPERATORS;

        case "BOOLEAN_COMPARISON_TYPE":
          return BOOLEAN_OPERATORS;
      }
    },
    hasValue(constraint) {
      if (constraint.type === "BOOLEAN_COMPARISON_TYPE") {
        return false;
      }

      return (
        constraint.operator !== "present" &&
        constraint.operator !== "notpresent" &&
        constraint.operator !== "empty" &&
        constraint.operator !== "notempty"
      );
    },
    getSegment() {
      let key = this.$route.params.key;

      Api.get("/segments/" + key)
        .then((response) => {
          this.segment = response.data;
          this.segment.constraints = response.data.constraints
            ? response.data.constraints
            : [];
        })
        .catch(() => {
          this.notifyError("Error loading segment.");
          this.$router.push("/segments");
        });
    },
    deleteSegment() {
      Api.delete("/segments/" + this.segment.key)
        .then(() => {
          this.notifySuccess("Segment deleted!");
          this.$router.push("/segments");
        })
        .catch((error) => {
          if (error.response && error.response.data) {
            this.notifyError(capitalize(error.response.data.message));
          } else {
            this.notifyError("Error deleting segment.");
            console.error(error);
          }
        });
    },
    updateSegment() {
      Api.put("/segments/" + this.segment.key, this.segment)
        .then(() => {
          this.notifySuccess("Segment updated!");
        })
        .catch((error) => {
          if (error.response && error.response.data) {
            this.notifyError(capitalize(error.response.data.message));
          } else {
            this.notifyError("Error updating segment.");
            console.error(error);
          }
        });
    },
    addConstraint() {
      Api.post(
        "/segments/" + this.segment.key + "/constraints",
        this.newConstraint
      )
        .then((response) => {
          this.segment.constraints.push(response.data);
          this.newConstraint = clone(DEFAULT_CONSTRAINT);
          this.notifySuccess("Constraint added!");
          this.dialogAddConstraintVisible = false;
        })
        .catch((error) => {
          if (error.response && error.response.data) {
            this.notifyError(capitalize(error.response.data.message));
          } else {
            this.notifyError("Error creating constraint.");
            console.error(error);
          }
        });
    },
    cancelAddConstraint() {
      this.dialogAddConstraintVisible = false;
      this.newConstraint = clone(DEFAULT_CONSTRAINT);
    },
    updateConstraint() {
      Api.put(
        "/segments/" +
          this.segment.key +
          "/constraints/" +
          this.selectedConstraint.id,
        this.selectedConstraint
      )
        .then((response) => {
          let constraint = response.data;
          let index = this.segment.constraints.findIndex(
            (c) => c.id === constraint.id
          );
          this.$set(this.segment.constraints, index, constraint);
          this.selectedConstraint = clone(DEFAULT_CONSTRAINT);
          this.notifySuccess("Constraint updated!");
          this.dialogEditConstraintVisible = false;
        })
        .catch((error) => {
          if (error.response && error.response.data) {
            this.notifyError(capitalize(error.response.data.message));
          } else {
            this.notifyError("Error updating constraint.");
            console.error(error);
          }
        });
    },
    deleteConstraint(index) {
      if (!confirm(`Are you sure you want to delete this constraint?`)) {
        return;
      }

      let constraint = this.segment.constraints[index];

      Api.delete(
        "/segments/" + this.segment.key + "/constraints/" + constraint.id
      )
        .then(() => {
          this.segment.constraints.splice(index, 1);
        })
        .catch((error) => {
          this.notifyError("Error deleting constraint.");
          console.error(error);
        });
    },
    editConstraint(index) {
      this.dialogEditConstraintVisible = true;
      this.selectedConstraint = cloneDeep(this.segment.constraints[index]);
    },
    duplicateConstraint(index) {
      this.dialogAddConstraintVisible = true;
      this.newConstraint = cloneDeep(this.segment.constraints[index]);
    },
    cancelEditConstraint() {
      this.dialogEditConstraintVisible = false;
      this.selectedConstraint = clone(DEFAULT_CONSTRAINT);
    },
  },
};
</script>
