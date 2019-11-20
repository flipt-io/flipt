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
            @input="formatKey"
          />
        </b-field>
        <b-field label="Description (optional)">
          <b-input
            v-model="segment.description"
            placeholder="Segment description"
          />
        </b-field>
        <hr />
        <div class="level">
          <div class="level-left">
            <div class="level-item">
              <div class="field is-grouped">
                <div class="control">
                  <button
                    class="button is-primary"
                    :disabled="!canCreate"
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

export default {
  name: "NewSegment",
  mixins: [notify, autoKeys],
  data() {
    return {
      segment: {}
    };
  },
  computed: {
    canCreate() {
      return this.segment.name && this.segment.key;
    }
  },
  methods: {
    formatKey() {
      this.segment.key = this.formatStringAsKey(this.segment.key);
    },
    setKeyIfSameAsName() {
      // Remove the character that was just added before comparing
      let prevName = this.segment.name.slice(0, -1);

      // Check if the name and key are currently in sync
      // We do this so we don't override a custom key value
      if (
        this.keyIsUndefinedOrEmpty() ||
        this.segment.key === this.formatStringAsKey(prevName)
      ) {
        this.segment.key = this.segment.name;
        this.formatKey();
      }
    },
    keyIsUndefinedOrEmpty() {
      return this.segment.key === undefined || this.segment.key === "";
    },
    createSegment() {
      Api.post("/segments", this.segment)
        .then(response => {
          this.notifySuccess("Segment created!");
          this.$router.push("/segments/" + response.data.key);
        })
        .catch(error => {
          if (error.response && error.response.data) {
            this.notifyError(capitalize(error.response.data.message));
          } else {
            this.notifyError("Error creating segment.");
            console.error(error);
          }
        });
    }
  }
};
</script>
