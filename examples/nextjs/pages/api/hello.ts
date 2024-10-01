// Next.js API route support: https://nextjs.org/docs/api-routes/introduction
import type { NextApiRequest, NextApiResponse } from "next";
import { FliptApiClient } from "@flipt-io/flipt";
import { v4 as uuidv4 } from "uuid";

const client = new FliptApiClient({
  environment: process.env.FLIPT_ADDR ?? "http://flipt:8080",
});

type Data = {
  name: string;
};

export default async function handler(
  _req: NextApiRequest,
  res: NextApiResponse<Data>,
) {
  let language = "english";
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

  let response: any = {
    greeting:
      language == "spanish"
        ? "Hola, from Next.js API route"
        : "Hello, from Next.js API route",
  };

  res.status(200).json(response);
}
