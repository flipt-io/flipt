import Greeting from "./components/Greeting";
import { getGreeting } from "./actions/getGreeting";
import { FliptProvider } from "./components/FliptProvider";
export default async function Home() {
  const { greeting } = await getGreeting();

  return (
    <main className="flex">
      <div className="flex flex-row justify-between w-full h-screen divide-x divide-gray-500 text-center">
        <div className="w-1/2 bg-black text-white flex items-center justify-center">
          <h1 className="text-3xl font-bold align-middle">{greeting}</h1>
        </div>
        <div className="w-1/2 flex items-center justify-center">
          <FliptProvider>
            <Greeting />
          </FliptProvider>
        </div>
      </div>
    </main>
  );
}
