<template>
  <section class="section">
    <div class="container">
      <div class="level">
        <div class="level-left">
          <div class="level-item">
            <div v-if="flags.length > 10" class="field has-addons">
              <p class="control">
                <input
                  v-model="search"
                  class="input"
                  type="text"
                  placeholder="Find a Flag"
                />
              </p>
              <p class="control"><button class="button">Search</button></p>
            </div>
          </div>
        </div>
        <div class="level-right">
          <div class="level-item">
            <RouterLink class="button is-primary" :to="{ name: 'new-flag' }"
              >New Flag</RouterLink
            >
          </div>
        </div>
      </div>
      <b-table
        :data="filteredFlags"
        :paginated="flags.length > 20"
        per-page="20"
        icon-pack="fas"
        hoverable="true"
      >
        <template slot-scope="props">
          <b-table-column field="enabled" label="Enabled">
            <span v-if="props.row.enabled" class="tag is-primary is-rounded"
              >On</span
            >
            <span v-else class="tag is-light is-rounded">Off</span>
          </b-table-column>
          <b-table-column field="key" label="Key" sortable>
            <RouterLink :to="{ name: 'flag', params: { key: props.row.key } }">
              {{ props.row.key }}
            </RouterLink>
          </b-table-column>
          <b-table-column field="name" label="Name" sortable>
            <RouterLink :to="{ name: 'flag', params: { key: props.row.key } }">
              {{ props.row.name }}
            </RouterLink>
          </b-table-column>
          <b-table-column field="hasVariants" label="Variants">
            {{ props.row.variants ? "yes" : "no" }}
          </b-table-column>
          <b-table-column field="description" label="Description">
            <small>{{ props.row.description | limit }}</small>
          </b-table-column>
          <b-table-column field="createdAt" label="Created" sortable>
            <small>{{ props.row.createdAt | moment("from", "now") }}</small>
          </b-table-column>
          <b-table-column field="updatedAt" label="Updated" sortable>
            <small>{{ props.row.updatedAt | moment("from", "now") }}</small>
          </b-table-column>
        </template>

        <template slot="empty">
          <section class="section">
            <div class="content has-text-grey has-text-centered">
              <p>
                No flags found. Create a
                <RouterLink :to="{ name: 'new-flag' }">New Flag</RouterLink>.
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
  name: "Flags",
  mixins: [notify],
  data() {
    return {
      search: "",
      flags: []
    };
  },
  computed: {
    filteredFlags() {
      return this.flags.filter(flag => {
        return (
          flag.name && flag.name.toLowerCase().match(this.search.toLowerCase())
        );
      });
    }
  },
  mounted() {
    this.getFlags();
  },
  methods: {
    getFlags() {
      Api.get("/flags")
        .then(response => {
          this.flags = response.data.flags ? response.data.flags : [];
        })
        .catch(error => {
          this.notifyError("Error loading flags.");
          console.error(error);
        });
    }
  }
};
</script>
