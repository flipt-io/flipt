// Next.js API route support: https://nextjs.org/docs/api-routes/introduction
import type { NextApiRequest, NextApiResponse } from "next";
import { FliptClient } from "@flipt-io/flipt";
import { v4 as uuidv4 } from "uuid";

const client = new FliptClient({
  url: process.env.NEXT_PUBLIC_FLIPT_ADDR ?? "http://flipt:8080",
});

type Data = {
  greeting: string;
};

export default async function handler(
  _req: NextApiRequest,
  res: NextApiResponse<Data>,
) {
  let language = "en";
  try {
    const evaluation = await client.evaluation.variant({
      namespaceKey: "default",
      flagKey: "language",
      entityId: uuidv4(),
      context: {},
    });

    language = evaluation.variantKey;
  } catch (err) {
    console.log(err);
  }

  let greeting = "Hello, from Next.js client-side";

  switch (language) {
    case "es":
      greeting = "Hola, from Next.js client-side";
      break;
    case "fr":
      greeting = "Bonjour, from Next.js client-side";
      break;
    default:
      greeting = "Hello, from Next.js client-side";
  }

  res.status(200).json({
    greeting,
  });
}
