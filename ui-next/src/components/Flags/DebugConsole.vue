<template>
  <div>
    <h5 class="title is-5">Debug Console</h5>
    <div class="columns">
      <div class="column">
        <h6 class="title is-6">Request</h6>
        <textarea
          class="textarea"
          :class="{ 'is-danger': invalidRequest }"
          rows="15"
          :value="JSON.stringify(request, undefined, 4)"
          @change="updateRequest"
        />
      </div>
      <div class="column">
        <h6 class="title is-6">Response</h6>
        <textarea
          class="textarea"
          :class="responseClass"
          rows="15"
          :value="JSON.stringify(response, undefined, 4)"
          disabled
        />
      </div>
    </div>
    <div class="field">
      <div class="control is-grouped">
        <button class="button is-primary" @click.prevent="evaluate()">
          Debug
        </button>
        <button class="button" @click.prevent="reset()">Reset</button>
      </div>
    </div>
  </div>
</template>

<script>
import { clone, isEmpty } from "lodash";
import { v4 as uuidv4 } from "uuid";

import { Api } from "@/services/api";

const DEFAULT_REQUEST = {
  flagKey: "",
  entityId: uuidv4(),
  context: {
    foo: "bar",
  },
};

export default {
  name: "FlagDebugConsole",
  props: {
    flag: Object,
  },
  data() {
    return {
      invalidRequest: false,
      request: clone(DEFAULT_REQUEST),
      response: {},
    };
  },
  computed: {
    responseClass() {
      if (isEmpty(this.response)) {
        return "";
      } else if (this.response.error) {
        return "is-danger";
      } else {
        return "is-success";
      }
    },
  },
  watch: {
    flag: function () {
      this.$set(this.request, "flagKey", this.flag.key);
    },
  },
  methods: {
    updateRequest(e) {
      try {
        this.request = JSON.parse(e.target.value);
        this.invalidRequest = false;
      } catch (error) {
        this.invalidRequest = true;
      }
    },
    reset() {
      this.request = clone(DEFAULT_REQUEST);
      this.request.entityId = uuidv4();
      this.request.flagKey = this.flag.key;
      this.response = {};
    },
    evaluate() {
      Api.post("/evaluate", this.request)
        .then((response) => {
          this.response = response.data;
        })
        .catch((error) => {
          this.response = { error: error.response.data.error };
        });
    },
  },
};
</script>
