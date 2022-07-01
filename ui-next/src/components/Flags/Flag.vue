<template>
  <div>
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
            <li class="is-active"><a>Details</a></li>
            <li>
              <RouterLink
                :to="{ name: 'targeting', params: { key: $route.params.key } }"
              >
                Targeting
              </RouterLink>
            </li>
          </ul>
        </div>
        <form>
          <b-field label="Key">
            <b-input v-model="flag.key" disabled />
          </b-field>
          <b-field label="Name">
            <b-input v-model="flag.name" placeholder="Flag name" required />
          </b-field>
          <b-field label="Description (optional)">
            <b-input
              v-model="flag.description"
              placeholder="Flag description"
            />
          </b-field>
          <b-field label="Enabled">
            <b-switch v-model="flag.enabled" />
          </b-field>
          <hr />
          <div class="level">
            <div class="level-left">
              <div class="level-item">
                <div class="field is-grouped">
                  <div class="control">
                    <button
                      class="button is-primary"
                      :disabled="!canUpdateFlag"
                      @click.prevent="updateFlag()"
                    >
                      Save
                    </button>
                  </div>
                  <div class="control">
                    <RouterLink class="button is-text" :to="{ name: 'flags' }">
                      Cancel
                    </RouterLink>
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
                      @click.prevent="dialogDeleteFlagVisible = true"
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
        <h5 class="title is-5">Variants</h5>
        <p class="subtitle is-7">
          Return different values based on rules you define
        </p>
        <b-table :data="flag.variants" scrollable>
          <b-table-column v-slot="props" field="key" label="Key" sortable>
            {{ props.row.key }}
          </b-table-column>
          <b-table-column v-slot="props" field="name" label="Name" sortable>
            {{ props.row.name }}
          </b-table-column>
          <b-table-column
            v-slot="props"
            field="description"
            label="Description"
            sortable
          >
            {{ props.row.description }}
          </b-table-column>
          <b-table-column
            v-slot="props"
            field="attachment"
            label="Attachment"
            sortable
          >
            {{ props.row.attachment | limit }}
          </b-table-column>
          <b-table-column v-slot="props" field="" label="" width="160" centered>
            <a
              class="button is-white"
              @click.prevent="editVariant(props.index)"
            >
              <span class="icon is-small">
                <i class="fas fa-pencil-alt" title="Edit" />
              </span>
            </a>
            <a
              class="button is-white"
              @click.prevent="duplicateVariant(props.index)"
            >
              <span class="icon is-small">
                <i class="far fa-clone" title="Duplicate" />
              </span>
            </a>
            <a
              class="button is-white"
              @click.prevent="deleteVariant(props.index)"
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
              @click.prevent="dialogAddVariantVisible = true"
            >
              New Variant
            </button>
          </div>
        </div>
      </div>
    </section>

    <div
      id="addVariantDialog"
      class="modal"
      tabindex="0"
      :class="{ 'is-active': dialogAddVariantVisible }"
      @keyup.esc="cancelAddVariant"
    >
      <div class="modal-background" @click.prevent="cancelAddVariant" />
      <div class="modal-content">
        <div class="container">
          <div class="box">
            <form>
              <b-field label="New Variant">
                <b-input
                  v-model="newVariant.key"
                  placeholder="Key"
                  required
                  @input="formatVariantKey(newVariant)"
                />
              </b-field>
              <b-field label="Name (optional)">
                <b-input v-model="newVariant.name" placeholder="Name" />
              </b-field>
              <b-field label="Description (optional)">
                <b-input
                  v-model="newVariant.description"
                  placeholder="Description"
                />
              </b-field>
              <b-field label="Attachment (optional)">
                <b-input
                  v-model="newVariant.attachment"
                  maxlength="1000"
                  type="textarea"
                  placeholder="{}"
                />
              </b-field>
              <div class="field is-grouped">
                <div class="control">
                  <button
                    class="button is-primary"
                    :disabled="!canAddVariant"
                    @click.prevent="addVariant"
                  >
                    Add Variant
                  </button>
                  <button
                    class="button is-text"
                    @click.prevent="cancelAddVariant"
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
            @click.prevent="cancelAddVariant"
          />
        </div>
      </div>
    </div>

    <div
      id="editVariantDialog"
      class="modal"
      tabindex="0"
      :class="{ 'is-active': dialogEditVariantVisible }"
      @keyup.esc="cancelEditVariant"
    >
      <div class="modal-background" @click.prevent="cancelEditVariant" />
      <div class="modal-content">
        <div class="container">
          <div class="box">
            <form>
              <b-field label="New Variant">
                <b-input
                  v-model="selectedVariant.key"
                  placeholder="Key"
                  required
                  @input="formatVariantKey(selectedVariant)"
                />
              </b-field>
              <b-field label="Name (optional)">
                <b-input v-model="selectedVariant.name" placeholder="Name" />
              </b-field>
              <b-field label="Description (optional)">
                <b-input
                  v-model="selectedVariant.description"
                  placeholder="Description"
                />
              </b-field>
              <b-field label="Attachment (optional)">
                <b-input
                  v-model="selectedVariant.attachment"
                  maxlength="1000"
                  type="textarea"
                  placeholder="{}"
                />
              </b-field>
              <div class="field is-grouped">
                <div class="control">
                  <button
                    class="button is-primary"
                    :disabled="!canUpdateVariant"
                    @click.prevent="updateVariant"
                  >
                    Update Variant
                  </button>
                  <button
                    class="button is-text"
                    @click.prevent="cancelEditVariant"
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
            @click.prevent="cancelEditVariant"
          />
        </div>
      </div>
    </div>

    <div
      id="deleteFlagDialog"
      class="modal"
      tabindex="0"
      :class="{ 'is-active': dialogDeleteFlagVisible }"
      @keyup.esc="dialogDeleteFlagVisible = false"
    >
      <div class="modal-background" @click="dialogDeleteFlagVisible = false" />
      <div class="modal-content">
        <div class="container">
          <div class="box">
            <p class="has-text-centered">
              Are you sure you want to delete this flag?
            </p>
            <br />
            <div class="control has-text-centered">
              <button class="button is-danger" @click.prevent="deleteFlag">
                Confirm
              </button>
              <button
                class="button is-text"
                @click.prevent="dialogDeleteFlagVisible = false"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
        <button
          class="modal-close is-large"
          aria-label="close"
          @click="dialogDeleteFlagVisible = false"
        />
      </div>
    </div>
  </div>
</template>

<script>
import { capitalize, clone, cloneDeep } from "lodash";

import { Api } from "@/services/api";
import notify from "@/mixins/notify";
import utils from "@/mixins/utils";

const DEFAULT_VARIANT = {
  key: "",
  name: "",
  description: "",
};

export default {
  name: "Flag",
  mixins: [notify, utils],
  data() {
    return {
      dialogDeleteFlagVisible: false,
      dialogAddVariantVisible: false,
      dialogEditVariantVisible: false,
      flag: {
        variants: [],
      },
      newVariant: clone(DEFAULT_VARIANT),
      selectedVariant: clone(DEFAULT_VARIANT),
    };
  },
  computed: {
    canUpdateFlag() {
      return this.isPresent(this.flag.key) && this.isPresent(this.flag.name);
    },
    canAddVariant() {
      return this.isPresent(this.newVariant.key);
    },
    canUpdateVariant() {
      return this.isPresent(this.selectedVariant.key);
    },
  },
  mounted() {
    this.getFlag();
  },
  methods: {
    formatVariantKey(variant) {
      variant.key = variant.key.toLowerCase().split(" ").join("-");
    },
    getFlag() {
      let key = this.$route.params.key;

      Api.get("/flags/" + key)
        .then((response) => {
          this.flag = response.data;
          this.flag.variants = response.data.variants
            ? response.data.variants
            : [];
        })
        .catch(() => {
          this.notifyError("Error loading flag.");
          this.$router.push("/flags");
        });
    },
    deleteFlag() {
      Api.delete("/flags/" + this.flag.key)
        .then(() => {
          this.notifySuccess("Flag deleted!");
          this.$router.push("/flags");
        })
        .catch((error) => {
          this.notifyError("Error deleting flag.");
          console.error(error);
        });
    },
    updateFlag() {
      Api.put("/flags/" + this.flag.key, this.flag)
        .then(() => {
          this.notifySuccess("Flag updated!");
        })
        .catch((error) => {
          if (error.response && error.response.data) {
            this.notifyError(capitalize(error.response.data.message));
          } else {
            this.notifyError("Error updating flag.");
            console.error(error);
          }
        });
    },
    addVariant() {
      Api.post("/flags/" + this.flag.key + "/variants", this.newVariant)
        .then((response) => {
          this.flag.variants.push(response.data);
          this.newVariant = clone(DEFAULT_VARIANT);
          this.notifySuccess("Variant added!");
          this.dialogAddVariantVisible = false;
        })
        .catch((error) => {
          if (error.response && error.response.data) {
            this.notifyError(capitalize(error.response.data.message));
          } else {
            this.notifyError("Error creating variant.");
            console.error(error);
          }
        });
    },
    cancelAddVariant() {
      this.dialogAddVariantVisible = false;
      this.newVariant = clone(DEFAULT_VARIANT);
    },
    updateVariant() {
      Api.put(
        "/flags/" + this.flag.key + "/variants/" + this.selectedVariant.id,
        this.selectedVariant
      )
        .then((response) => {
          let variant = response.data;
          let index = this.flag.variants.findIndex((v) => v.id === variant.id);
          this.$set(this.flag.variants, index, variant);
          this.selectedVariant = clone(DEFAULT_VARIANT);
          this.notifySuccess("Variant updated!");
          this.dialogEditVariantVisible = false;
        })
        .catch((error) => {
          if (error.response && error.response.data) {
            this.notifyError(capitalize(error.response.data.message));
          } else {
            this.notifyError("Error updating variant.");
            console.error(error);
          }
        });
    },
    deleteVariant(index) {
      if (!confirm(`Are you sure you want to delete this variant?`)) {
        return;
      }

      let variant = this.flag.variants[index];

      Api.delete("/flags/" + this.flag.key + "/variants/" + variant.id)
        .then(() => {
          this.flag.variants.splice(index, 1);
          this.notifySuccess("Variant deleted!");
        })
        .catch((error) => {
          if (error.response && error.response.data) {
            this.notifyError(capitalize(error.response.data.message));
          } else {
            this.notifyError("Error deleting variant.");
            console.error(error);
          }
        });
    },
    editVariant(index) {
      this.dialogEditVariantVisible = true;
      this.selectedVariant = cloneDeep(this.flag.variants[index]);
    },
    duplicateVariant(index) {
      this.dialogAddVariantVisible = true;
      this.newVariant = cloneDeep(this.flag.variants[index]);
    },
    cancelEditVariant() {
      this.dialogEditVariantVisible = false;
      this.selectedVariant = clone(DEFAULT_VARIANT);
    },
  },
};
</script>
