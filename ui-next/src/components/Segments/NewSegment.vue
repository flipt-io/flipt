<template>
  <section class="section">
    <div class="container">
      <nav class="breadcrumb" aria-label="breadcrumbs">
        <ul>
          <li><RouterLink :to="{ name: 'segments' }">Segments</RouterLink></li>
          <li class="is-active">
            <a href="#" aria-current="page">New Segment</a>
          </li>
        </ul>
      </nav>
      <form>
        <b-field label="Name">
          <b-input
            v-model="segment.name"
            placeholder="Segment name"
            required
            @input="setKeyIfSameAsName"
          />
        </b-field>
        <b-field label="Key">
          <b-input
            v-model="segment.key"
            placeholder="Segment key"
            required
            validation-message="Only letters, numbers, hypens and underscores allowed"
            pattern="^[-_,A-Za-z0-9]+$"
            @input="formatKey"
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
            <b-radio v-model="segment.matchType" native-value="ALL_MATCH_TYPE">
              Match All
            </b-radio>
            <b-radio v-model="segment.matchType" native-value="ANY_MATCH_TYPE">
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
                    data-testid="create-segment"
                    class="button is-primary"
                    :disabled="!canCreateSegment"
                    @click.prevent="createSegment()"
                  >
                    Create
                  </button>
                </div>
                <div class="control">
                  <RouterLink class="button is-text" :to="{ name: 'segments' }"
                    >Cancel</RouterLink
                  >
                </div>
              </div>
            </div>
          </div>
        </div>
      </form>
    </div>
  </section>
</template>

<script>
import capitalize from "lodash/capitalize";
import { Api } from "@/services/api";
import notify from "@/mixins/notify";
import autoKeys from "@/mixins/autoKeys";
import utils from "@/mixins/utils";

export default {
  name: "NewSegment",
  mixins: [notify, autoKeys, utils],
  data() {
    return {
      isValid: false,
      segment: {
        matchType: "ALL_MATCH_TYPE",
      },
    };
  },
  computed: {
    canCreateSegment() {
      return (
        this.isPresent(this.segment.name) &&
        this.isPresent(this.segment.key) &&
        this.isValid
      );
    },
    matchTypeText() {
      if (this.segment.matchType === "ALL_MATCH_TYPE") {
        return "All constraints must match.";
      } else {
        return "At least one constraint must match.";
      }
    },
  },
  methods: {
    formatKey() {
      this.segment.key = this.formatStringAsKey(this.segment.key);
      this.isValid = this.segment.key.match("^[-_,A-Za-z0-9]+$");
    },
    setKeyIfSameAsName() {
      // Remove the character that was just added before comparing
      let prevName = this.segment.name.slice(0, -1);

      // Check if the name and key are currently in sync
      // We do this so we don't override a custom key value
      if (
        this.isEmpty(this.segment.key) ||
        this.segment.key === this.formatStringAsKey(prevName)
      ) {
        this.segment.key = this.segment.name;
        this.formatKey();
      }
    },
    createSegment() {
      Api.post("/segments", this.segment)
        .then((response) => {
          this.notifySuccess("Segment created!");
          this.$router.push("/segments/" + response.data.key);
        })
        .catch((error) => {
          if (error.response && error.response.data) {
            this.notifyError(capitalize(error.response.data.message));
          } else {
            this.notifyError("Error creating segment.");
            console.error(error);
          }
        });
    },
  },
};
</script>
