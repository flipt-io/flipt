"use server";

import { FliptClient } from "@flipt-io/flipt-client-js";
import { v4 as uuidv4 } from "uuid";
export async function getGreeting() {
  const flipt = await FliptClient.init({
    url: process.env.FLIPT_ADDR ?? "http://localhost:8080",
    updateInterval: 10,
  });

  const uuid = uuidv4();

  const result = flipt.evaluateVariant({
    flagKey: "language",
    entityId: uuid,
    context: {
      user_id: uuid,
    },
  });

  let language = result.variantKey;
  let greeting = "";

  switch (language) {
    case "es":
      greeting = "Hola, from Next.js server-side";
      break;
    case "fr":
      greeting = "Bonjour, from Next.js server-side";
      break;
    default:
      greeting = "Hello, from Next.js server-side";
  }

  return {
    greeting,
  };
}
