import { useState, useEffect } from "react";
import { v4 as uuidv4 } from "uuid";
import { FliptEvaluationClient } from "@flipt-io/flipt-client-browser";

export default function Greeting() {
  const [data, setData] = useState<any | null>(null);

  useEffect(() => {
    async function fetchData() {
      try {
        const client = await FliptEvaluationClient.init("default", {
          url: process.env.NEXT_PUBLIC_FLIPT_ADDR ?? "http://localhost:8080",
        });

        const result = client.evaluateVariant("language", uuidv4(), {});
        console.log(result);

        let language = result.variantKey;

        const greeting =
          language == "es"
            ? "Hola, from Next.js client-side"
            : "Bonjour, from Next.js client-side";

        setData(greeting);
      } catch (err) {
        console.log(err);
      }
    }

    fetchData();
  });

  if (!data)
    return <p className="text-xl font-bold align-middle"> Loading... </p>;

  return <h1 className="text-3xl font-bold align-middle">{data}</h1>;
}
