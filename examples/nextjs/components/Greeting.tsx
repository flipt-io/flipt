import { useState, useEffect } from "react";
import { useFlipt } from "../hooks/flipt";
import { v4 as uuidv4 } from "uuid";

export default function Greeting() {
  const [data, setData] = useState<any | null>(null);

  const flipt = useFlipt();

  useEffect(() => {
    async function fetchData() {
      let language = "french";
      try {
        const evaluation = await flipt.evaluate.evaluate({
          flagKey: "language",
          entityId: uuidv4(),
          context: {},
        });

        if (!evaluation.ok) {
          console.log(evaluation.error);
        } else {
          language = evaluation.body.value;
        }

        let greeting =
          language == "spanish"
            ? "Hola, from Next.js client-side"
            : "Bonjour, from Next.js client-side";

        setData(greeting);
      } catch (err) {
        console.log(err);
      }
    }

    fetchData();
  }, [flipt.evaluate]);

  if (!data)
    return <p className="text-xl font-bold align-middle"> Loading... </p>;

  return <h1 className="text-3xl font-bold align-middle">{data}</h1>;
}
