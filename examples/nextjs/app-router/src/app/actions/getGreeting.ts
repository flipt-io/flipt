"use server";

import { FliptClient } from "@flipt-io/flipt";
import { v4 as uuidv4 } from "uuid";

export async function getGreeting() {
  const flipt = new FliptClient({
    url: process.env.FLIPT_ADDR ?? "http://flipt:8080",
  });

  const uuid = uuidv4();

  const result = await flipt.evaluation.variant({
    namespaceKey: "default",
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
