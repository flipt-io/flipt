import { FliptApi, FliptApiClient } from "@flipt-io/flipt";
import React from "react";

const FliptContext = React.createContext<FliptApiClient | null>(null);

export const FliptProvider = ({ children }: { children: React.ReactNode }) => {
  const client = new FliptApiClient({
    environment: process.env.FLIPT_PUBLIC_ADDR ?? "http://localhost:8081",
  });
  return (
    <FliptContext.Provider value={client}>{children}</FliptContext.Provider>
  );
};

export const useFlipt = () => {
  const client = React.useContext(FliptContext);
  if (!client) {
    throw new Error("useFlipt must be used within a FliptProvider");
  }
  return client;
};
