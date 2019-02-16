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
        <BField label="Key">
          <BInput
            v-model="segment.key"
            placeholder="Segment key"
            required
            @input="formatKey"
          />
        </BField>
        <BField label="Name">
          <BInput v-model="segment.name" placeholder="Segment name" required />
        </BField>
        <BField label="Description (optional)">
          <BInput
            v-model="segment.description"
            placeholder="Segment description"
          />
        </BField>
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

export default {
  name: "NewSegment",
  mixins: [notify],
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
      this.segment.key = this.segment.key
        .toLowerCase()
        .split(" ")
        .join("-");
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
