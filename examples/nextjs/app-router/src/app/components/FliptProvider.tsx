"use client";

import { FliptProvider as FliptProviderReact } from "@flipt-io/flipt-client-react";

export function FliptProvider({ children }: { children: React.ReactNode }) {
  return (
    <FliptProviderReact
      namespace="default"
      options={{
        url: process.env.NEXT_PUBLIC_FLIPT_ADDR ?? "http://flipt:8080",
      }}
    >
      {children}
    </FliptProviderReact>
  );
}
