import { FliptEvaluationClient } from "@flipt-io/flipt-client";
import Head from "next/head";
import Greeting from "../components/Greeting";
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
            <Greeting />
          </div>
        </div>
      </main>
    </>
  );
}


export async function getServerSideProps() {
  const client = await FliptEvaluationClient.init("default", {url: process.env.FLIPT_ADDR ?? "http://flipt:8080"}  );
  let language = "en";
  try {
    const result = client.evaluateVariant("language", uuidv4(), {});
    language = result.variantKey;
  } catch (err) {
    console.log(err);
  }

  const greeting =
    language == "es"
      ? "Hola, from Next.js server-side"
      : "Hello, from Next.js server-side";

  return {
    props: {
      greeting,
    },
  };
}
