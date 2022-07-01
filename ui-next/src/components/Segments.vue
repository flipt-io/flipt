<template>
  <section class="section">
    <div class="container">
      <div class="level">
        <div class="level-left">
          <div class="level-item">
            <div v-if="segments.length > 10" class="field has-addons">
              <p class="control">
                <input
                  v-model="search"
                  class="input"
                  type="text"
                  placeholder="Find a Segment"
                />
              </p>
              <p class="control"><button class="button">Search</button></p>
            </div>
          </div>
        </div>
        <div class="level-right">
          <div class="level-item">
            <RouterLink
              data-testid="new-segment"
              class="button is-primary"
              :to="{ name: 'new-segment' }"
            >
              New Segment
            </RouterLink>
          </div>
        </div>
      </div>
      <b-table
        :data="filteredSegments"
        :paginated="segments.length > 20"
        per-page="20"
        icon-pack="fas"
        hoverable="true"
      >
        <b-table-column v-slot="props" field="key" label="Key" sortable>
          <RouterLink :to="{ name: 'segment', params: { key: props.row.key } }">
            {{ props.row.key }}
          </RouterLink>
        </b-table-column>
        <b-table-column v-slot="props" field="name" label="Name" sortable>
          <RouterLink :to="{ name: 'segment', params: { key: props.row.key } }">
            {{ props.row.name }}
          </RouterLink>
        </b-table-column>
        <b-table-column
          v-slot="props"
          field="hasConstraints"
          label="Constraints"
        >
          {{ props.row.constraints.length > 0 ? "yes" : "no" }}
        </b-table-column>
        <b-table-column v-slot="props" field="hasConstraints" label="Match">
          {{ props.row.matchType === "ANY_MATCH_TYPE" ? "any" : "all" }}
        </b-table-column>
        <b-table-column v-slot="props" field="description" label="Description">
          <small>{{ props.row.description | limit }}</small>
        </b-table-column>
        <b-table-column
          v-slot="props"
          field="createdAt"
          label="Created"
          sortable
        >
          <small>{{ props.row.createdAt | moment("from", "now") }}</small>
        </b-table-column>
        <b-table-column
          v-slot="props"
          field="updatedAt"
          label="Updated"
          sortable
        >
          <small>{{ props.row.updatedAt | moment("from", "now") }}</small>
        </b-table-column>

        <template #empty>
          <section class="section">
            <div class="content has-text-grey has-text-centered">
              <p>
                No segments found. Create a
                <RouterLink :to="{ name: 'new-segment' }">
                  New Segment</RouterLink
                >.
              </p>
            </div>
          </section>
        </template>
      </b-table>
    </div>
  </section>
</template>

<script>
import { Api } from "@/services/api";
import notify from "@/mixins/notify";

export default {
  name: "Segments",
  mixins: [notify],
  data() {
    return {
      search: "",
      segments: [],
    };
  },
  computed: {
    filteredSegments() {
      return this.segments.filter((segment) => {
        return segment.name.toLowerCase().match(this.search.toLowerCase());
      });
    },
  },
  mounted() {
    this.getSegments();
  },
  methods: {
    getSegments() {
      Api.get("/segments")
        .then((response) => {
          this.segments = response.data.segments ? response.data.segments : [];
        })
        .catch((error) => {
          this.notifyError("Error loading segments.");
          console.error(error);
        });
    },
  },
};
</script>
