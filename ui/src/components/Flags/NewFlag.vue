<template>
  <section class="section">
    <div class="container">
      <nav class="breadcrumb" aria-label="breadcrumbs">
        <ul>
          <li><RouterLink :to="{ name: 'flags' }">Flags</RouterLink></li>
          <li class="is-active">
            <a href="#" aria-current="page">New Flag</a>
          </li>
        </ul>
      </nav>
      <div class="tabs is-boxed">
        <ul>
          <li class="is-active"><a>Details</a></li>
        </ul>
      </div>
      <form>
        <BField label="Name">
          <BInput
            v-model="flag.name"
            placeholder="Flag name"
            required
            @input="setKeyIfSameAsName"
          />
        </BField>
        <BField label="Key">
          <BInput
            v-model="flag.key"
            placeholder="Flag key"
            required
            @input="formatKey"
          />
        </BField>
        <BField label="Description (optional)">
          <BInput v-model="flag.description" placeholder="Flag description" />
        </BField>
        <BField label="Enabled"> <BSwitch v-model="flag.enabled" /> </BField>
        <hr />
        <div class="level">
          <div class="level-left">
            <div class="level-item">
              <div class="field is-grouped">
                <div class="control">
                  <button
                    class="button is-primary"
                    :disabled="!canCreate"
                    @click.prevent="createFlag()"
                  >
                    Create
                  </button>
                </div>
                <div class="control">
                  <RouterLink class="button is-text" :to="{ name: 'flags' }"
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
  name: "NewFlag",
  mixins: [notify],
  data() {
    return {
      flag: {}
    };
  },
  computed: {
    canCreate() {
      return this.flag.name && this.flag.key;
    }
  },
  methods: {
    formatKey() {
      this.flag.key = this.formatStringAsKey(this.flag.key);
    },
    formatStringAsKey(str) {
      return str
        .toLowerCase()
        .split(" ")
        .join("-");
    },
    setKeyIfSameAsName() {
      // Remove the character that was just added before comparing
      let prevName = this.flag.name.slice(0, -1);

      // Check if the name and key are currently in sync
      // We do this so we don't override a custom key value
      if (
        this.keyIsUndefinedOrEmpty() ||
        this.flag.key === this.formatStringAsKey(prevName)
      ) {
        this.flag.key = this.flag.name;
        this.formatKey();
      }
    },
    keyIsUndefinedOrEmpty() {
      return this.flag.key === undefined || this.flag.key === "";
    },
    createFlag() {
      Api.post("/flags", this.flag)
        .then(response => {
          this.notifySuccess("Flag created!");
          this.$router.push("/flags/" + response.data.key);
        })
        .catch(error => {
          if (error.response && error.response.data) {
            this.notifyError(capitalize(error.response.data.message));
          } else {
            this.notifyError("Error creating flag.");
            console.error(error);
          }
        });
    }
  }
};
</script>
