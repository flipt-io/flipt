import { useState, useEffect } from "react";
import { v4 as uuidv4 } from "uuid";
import { useFliptContext } from "@flipt-io/flipt-client-react";

export default function Greeting() {
  const [data, setData] = useState<any | null>(null);

  const { client } = useFliptContext();

  useEffect(() => {
    async function fetchData() {
      try {
        const evaluation = client?.evaluateVariant("language", uuidv4(), {});
        console.log(evaluation);

        let language = evaluation?.variantKey;

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
