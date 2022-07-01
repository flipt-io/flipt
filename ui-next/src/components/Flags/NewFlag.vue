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
      <form ref="form">
        <b-field label="Name">
          <b-input
            v-model="flag.name"
            placeholder="Flag name"
            required
            @input="setKeyIfSameAsName"
          />
        </b-field>
        <b-field label="Key">
          <b-input
            v-model="flag.key"
            placeholder="Flag key"
            required
            validation-message="Only letters, numbers, hypens and underscores allowed"
            pattern="^[-_,A-Za-z0-9]+$"
            @input="formatKey"
          />
        </b-field>
        <b-field label="Description (optional)">
          <b-input v-model="flag.description" placeholder="Flag description" />
        </b-field>
        <b-field label="Enabled"> <b-switch v-model="flag.enabled" /> </b-field>
        <hr />
        <div class="level">
          <div class="level-left">
            <div class="level-item">
              <div class="field is-grouped">
                <div class="control">
                  <button
                    data-testid="create-flag"
                    class="button is-primary"
                    :disabled="!canCreateFlag"
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
import autoKeys from "@/mixins/autoKeys";
import utils from "@/mixins/utils";

export default {
  name: "NewFlag",
  mixins: [notify, autoKeys, utils],
  data() {
    return {
      isValid: false,
      flag: {},
    };
  },
  computed: {
    canCreateFlag() {
      return (
        this.isPresent(this.flag.name) &&
        this.isPresent(this.flag.key) &&
        this.isValid
      );
    },
  },
  methods: {
    formatKey() {
      this.flag.key = this.formatStringAsKey(this.flag.key);
      this.isValid = this.flag.key.match("^[-_,A-Za-z0-9]+$");
    },
    setKeyIfSameAsName() {
      // Remove the character that was just added before comparing
      let prevName = this.flag.name.slice(0, -1);

      // Check if the name and key are currently in sync
      // We do this so we don't override a custom key value
      if (
        this.isEmpty(this.flag.key) ||
        this.flag.key === this.formatStringAsKey(prevName)
      ) {
        this.flag.key = this.flag.name;
        this.formatKey();
      }
    },
    createFlag() {
      Api.post("/flags", this.flag)
        .then((response) => {
          this.notifySuccess("Flag created!");
          this.$router.push("/flags/" + response.data.key);
        })
        .catch((error) => {
          if (error.response && error.response.data) {
            this.notifyError(capitalize(error.response.data.message));
          } else {
            this.notifyError("Error creating flag.");
            console.error(error);
          }
        });
    },
  },
};
</script>
