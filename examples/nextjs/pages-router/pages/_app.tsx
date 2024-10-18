import "../styles/globals.css";
import type { AppProps } from "next/app";
import { FliptProvider } from "@flipt-io/flipt-client-react";

export default function App({ Component, pageProps }: AppProps) {
  return (
    <FliptProvider
      namespace="default"
      options={{
        url: process.env.NEXT_PUBLIC_FLIPT_ADDR ?? "http://flipt:8080",
      }}
    >
      <Component {...pageProps} />
    </FliptProvider>
  );
}
