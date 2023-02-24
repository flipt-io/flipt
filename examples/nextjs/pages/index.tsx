import { FliptApiClient } from "@flipt-io/flipt";
import Head from "next/head";
import Greeting from "../components/Greeting";
import { FliptProvider } from "../hooks/flipt";
import { v4 as uuidv4 } from "uuid";

type HomeProps = {
  greeting: string;
};

export default function Home(data: HomeProps) {
  return (
    <>
      <Head>
        <title>Example Flipt Integration with Next.js</title>
        <link rel="icon" href="/favicon.ico" />
      </Head>
      <main className="flex">
        <div className="flex flex-row justify-between w-full h-screen divide-x divide-gray-500 text-center">
          <div className="w-1/2 bg-black text-white flex items-center justify-center">
            <h1 className="text-3xl font-bold align-middle">{data.greeting}</h1>
          </div>
          <div className="w-1/2 flex items-center justify-center">
            <FliptProvider>
              <Greeting />
            </FliptProvider>
          </div>
        </div>
      </main>
    </>
  );
}

const client = new FliptApiClient({
  environment: process.env.FLIPT_ADDR ?? "http://localhost:8080",
});

export async function getServerSideProps() {
  let language = "english";
  try {
    const evaluation = await client.evaluate.evaluate({
      flagKey: "language",
      entityId: uuidv4(),
      context: {},
    });

    if (!evaluation.ok) {
      console.log(evaluation.error);
    } else {
      language = evaluation.body.value;
    }
  } catch (err) {
    console.log(err);
  }

  const greeting =
    language == "spanish"
      ? "Hola, from Next.js server-side"
      : "Hello, from Next.js server-side";

  return {
    props: {
      greeting,
    },
  };
}
