<template>
  <section class="section">
    <div class="container">
      <div class="level">
        <div class="level-left">
          <div class="level-item">
            <div v-if="flags.length > 5" class="field has-addons">
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
      <BTable :data="filteredFlags">
        <template slot-scope="props">
          <BTableColumn field="enabled" label="Enabled">
            <span v-if="props.row.enabled" class="tag is-primary is-rounded"
              >On</span
            >
            <span v-else class="tag is-light is-rounded">Off</span>
          </BTableColumn>
          <BTableColumn field="key" label="Key" sortable>
            <RouterLink :to="{ name: 'flag', params: { key: props.row.key } }">
              {{ props.row.key }}
            </RouterLink>
          </BTableColumn>
          <BTableColumn field="name" label="Name" sortable>
            <RouterLink :to="{ name: 'flag', params: { key: props.row.key } }">
              {{ props.row.name }}
            </RouterLink>
          </BTableColumn>
          <BTableColumn field="hasVariants" label="Variants" sortable>
            {{ props.row.variants ? "yes" : "no" }}
          </BTableColumn>
          <BTableColumn field="description" label="Description">
            <small>{{ props.row.description }}</small>
          </BTableColumn>
          <BTableColumn field="createdAt" label="Created" sortable>
            <small>{{ props.row.createdAt | moment("from", "now") }}</small>
          </BTableColumn>
          <BTableColumn field="updatedAt" label="Updated" sortable>
            <small>{{ props.row.updatedAt | moment("from", "now") }}</small>
          </BTableColumn>
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
      </BTable>
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
