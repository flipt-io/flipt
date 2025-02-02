import http from "k6/http";
import { check, group } from "k6";
import { uuidv4 } from "https://jslib.k6.io/k6-utils/1.4.0/index.js";

export const options = {
  // A number specifying the number of VUs to run concurrently.
  vus: 10,
  // A string specifying the total duration of the test run.
  scenarios: {
    default: {
      executor: "constant-arrival-rate",
      rate: 1000,
      timeUnit: "1s",
      duration: "60s",
      preAllocatedVUs: 1000,
    },
  },
  // The following section contains configuration options for execution of this
  // test script in Grafana Cloud.
  //
  // See https://grafana.com/docs/grafana-cloud/k6/get-started/run-cloud-tests-from-the-cli/
  // to learn about authoring and running k6 test scripts in Grafana k6 Cloud.
  //
  // cloud: {
  //   // The ID of the project to which the test is assigned in the k6 Cloud UI.
  //   // By default tests are executed in default project.
  //   projectID: "",
  //   // The name of the test in the k6 Cloud UI.
  //   // Test runs with the same name will be grouped.
  //   name: "loadtest.js"
  // },
};

const FLIPT_ADDR = __ENV.FLIPT_ADDR || "http://flipt:8080";
const FLIPT_AUTH_TOKEN = __ENV.FLIPT_AUTH_TOKEN || "";

const headers = {
  "Content-Type": "application/json",
  Authorization: `Bearer ${FLIPT_AUTH_TOKEN}`,
};

// The function that defines VU logic.
//
// See https://grafana.com/docs/k6/latest/examples/get-started-with-k6/ to learn more
// about authoring k6 scripts.
//
export default function () {
  group("getFlag", () => {
    // Test GET flag endpoint
    const flagResponse = http.get(
      `${FLIPT_ADDR}/api/v1/namespaces/default/flags/flag_001`,
      { headers },
    );
    check(flagResponse, {
      "flag status is 200": (r) => r.status === 200,
    });
  });

  group("evaluateVariant", () => {
    // Test POST variant evaluation endpoint
    const variantEvaluation = {
      entityId: uuidv4(),
      flagKey: "flag_010",
      context: {
        in_segment: "baz",
      },
    };

    const variantResponse = http.post(
      `${FLIPT_ADDR}/evaluate/v1/variant`,
      JSON.stringify(variantEvaluation),
      { headers },
    );
    check(variantResponse, {
      "variant eval status is 200": (r) => r.status === 200,
    });
  });

  group("evaluateBoolean", () => {
    // Test POST boolean evaluation endpoint
    const booleanEvaluation = {
      entityId: uuidv4(),
      flagKey: "flag_boolean",
      context: {
        in_segment: "baz",
      },
    };

    const booleanResponse = http.post(
      `${FLIPT_ADDR}/evaluate/v1/boolean`,
      JSON.stringify(booleanEvaluation),
      { headers },
    );
    check(booleanResponse, {
      "boolean eval status is 200": (r) => r.status === 200,
    });
  });
}
