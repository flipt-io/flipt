"use client";

import { useState, useEffect } from "react";
import { v4 as uuidv4 } from "uuid";
import { useFliptVariant } from "@flipt-io/flipt-client-react";

export default function Greeting() {
  const [uuid, setUuid] = useState(uuidv4());
  const [data, setData] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const language = useFliptVariant("language", "en", uuid, {
    user_id: uuid,
  });

  useEffect(() => {
    let greeting = "";

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

    setData(greeting);
  }, [language]);

  const handleRefresh = () => {
    setIsLoading(true);
    setUuid(uuidv4());
    setTimeout(() => setIsLoading(false), 100); // Simulate a delay
  };

  if (!data)
    return <p className="text-xl font-bold align-middle"> Loading... </p>;

  return (
    <div className="flex flex-col items-center justify-center space-y-4">
      <h1 className="text-3xl font-bold align-middle">{data}</h1>
      <button
        onClick={handleRefresh}
        className="mt-4 px-4 py-2 bg-black text-white rounded hover:bg-gray-800 flex items-center"
        disabled={isLoading}
      >
        {isLoading ? (
          <>
            <svg className="animate-spin h-5 w-5 mr-3" viewBox="0 0 24 24">
              <circle
                className="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                strokeWidth="4"
                fill="none"
              />
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              />
            </svg>
            Refreshing...
          </>
        ) : (
          "Refresh"
        )}
      </button>
    </div>
  );
}
