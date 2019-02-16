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
          <BField label="Key"> <BInput v-model="flag.key" disabled /> </BField>
          <BField label="Name">
            <BInput v-model="flag.name" placeholder="Flag name" required />
          </BField>
          <BField label="Description (optional)">
            <BInput v-model="flag.description" placeholder="Flag description" />
          </BField>
          <BField label="Enabled"> <BSwitch v-model="flag.enabled" /> </BField>
          <br />
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
                    <RouterLink class="button is-text" :to="{ name: 'flags' }"
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
        <hr />
        <h5 class="title is-5">Variants</h5>
        <p class="subtitle is-7">
          Return different values based on rules you define
        </p>
        <BTable :data="flag.variants">
          <template slot-scope="props">
            <BTableColumn field="key" label="Key" sortable>
              {{ props.row.key }}
            </BTableColumn>
            <BTableColumn field="name" label="Name" sortable>
              {{ props.row.name }}
            </BTableColumn>
            <BTableColumn field="description" label="Description" sortable>
              {{ props.row.description }}
            </BTableColumn>
            <BTableColumn field="" label="" width="110" centered>
              <a
                class="button is-white"
                @click.prevent="editVariant(props.index)"
              >
                <span class="icon is-small">
                  <i class="fas fa-pencil-alt" />
                </span>
              </a>
              <a
                class="button is-white"
                @click.prevent="deleteVariant(props.index)"
              >
                <span class="icon is-small"> <i class="fas fa-times" /> </span>
              </a>
            </BTableColumn>
          </template>
        </BTable>
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
      :class="{ 'is-active': dialogAddVariantVisible }"
    >
      <div class="modal-background" @click.prevent="cancelAddVariant" />
      <div class="modal-content" @keyup.esc="cancelAddVariant">
        <div class="box">
          <form>
            <BField label="New Variant">
              <BInput
                v-model="newVariant.key"
                placeholder="Key"
                required
                @input="formatVariantKey(newVariant)"
              />
            </BField>
            <BField label="Name (optional)">
              <BInput v-model="newVariant.name" placeholder="Name" />
            </BField>
            <BField label="Description (optional)">
              <BInput
                v-model="newVariant.description"
                placeholder="Description"
              />
            </BField>
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

    <div
      id="editVariantDialog"
      class="modal"
      :class="{ 'is-active': dialogEditVariantVisible }"
    >
      <div class="modal-background" @click.prevent="cancelEditVariant" />
      <div class="modal-content" @keyup.esc="cancelEditVariant">
        <div class="box">
          <form>
            <BField label="New Variant">
              <BInput
                v-model="selectedVariant.key"
                placeholder="Key"
                required
                @input="formatVariantKey(selectedVariant)"
              />
            </BField>
            <BField label="Name (optional)">
              <BInput v-model="selectedVariant.name" placeholder="Name" />
            </BField>
            <BField label="Description (optional)">
              <BInput
                v-model="selectedVariant.description"
                placeholder="Description"
              />
            </BField>
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

    <div
      id="deleteFlagDialog"
      class="modal"
      :class="{ 'is-active': dialogDeleteFlagVisible }"
    >
      <div class="modal-background" @click="dialogDeleteFlagVisible = false" />
      <div class="modal-content" @keyup.esc="dialogDeleteFlagVisible = false">
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
</template>

<script>
import clone from "lodash/clone";
import cloneDeep from "lodash/cloneDeep";
import capitalize from "lodash/capitalize";

import { Api } from "@/services/api";
import notify from "@/mixins/notify";

const DEFAULT_VARIANT = {
  key: "",
  name: "",
  description: ""
};

export default {
  name: "Flag",
  mixins: [notify],
  data() {
    return {
      dialogDeleteFlagVisible: false,
      dialogAddVariantVisible: false,
      dialogEditVariantVisible: false,
      flag: {
        variants: []
      },
      newVariant: clone(DEFAULT_VARIANT),
      selectedVariant: clone(DEFAULT_VARIANT)
    };
  },
  computed: {
    canUpdateFlag() {
      return this.flag.key && this.flag.name;
    },
    canAddVariant() {
      return this.newVariant.key;
    },
    canUpdateVariant() {
      return this.selectedVariant.key;
    }
  },
  mounted() {
    this.getFlag();
  },
  methods: {
    formatVariantKey(variant) {
      variant.key = variant.key
        .toLowerCase()
        .split(" ")
        .join("-");
    },
    getFlag() {
      let key = this.$route.params.key;

      Api.get("/flags/" + key)
        .then(response => {
          this.flag = response.data;
          this.flag.variants = response.data.variants
            ? response.data.variants
            : [];
        })
        .catch(error => {
          this.notifyError("Error loading flag.");
          console.error(error);
        });
    },
    deleteFlag() {
      Api.delete("/flags/" + this.flag.key)
        .then(() => {
          this.notifySuccess("Flag deleted!");
          this.$router.push("/flags");
        })
        .catch(error => {
          this.notifyError("Error deleting flag.");
          console.error(error);
        });
    },
    updateFlag() {
      Api.put("/flags/" + this.flag.key, this.flag)
        .then(() => {
          this.notifySuccess("Flag updated!");
        })
        .catch(error => {
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
        .then(response => {
          this.flag.variants.push(response.data);
          this.newVariant = clone(DEFAULT_VARIANT);
          this.notifySuccess("Variant added!");
          this.dialogAddVariantVisible = false;
        })
        .catch(error => {
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
        .then(response => {
          let variant = response.data;
          let index = this.flag.variants.findIndex(v => v.id === variant.id);
          this.flag.variants[index] = variant;
          this.selectedVariant = clone(DEFAULT_VARIANT);
          this.notifySuccess("Variant updated!");
          this.dialogEditVariantVisible = false;
        })
        .catch(error => {
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
        .catch(error => {
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
    cancelEditVariant() {
      this.dialogEditVariantVisible = false;
      this.selectedVariant = clone(DEFAULT_VARIANT);
    }
  }
};
</script>
