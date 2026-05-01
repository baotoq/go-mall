"use client";

import { useState } from "react";
import { Database } from "lucide-react";
import { Button } from "@/components/ui/button";
import { seedData } from "@/lib/api";

type State = "idle" | "loading" | "success" | "error";

export function SeedButton() {
  const [state, setState] = useState<State>("idle");

  async function handleSeed() {
    if (state === "loading") return;
    setState("loading");
    const result = await seedData();
    if (result === null) {
      setState("error");
    } else {
      setState("success");
    }
    setTimeout(() => setState("idle"), 3000);
  }

  return (
    <Button
      variant="outline"
      size="sm"
      onClick={handleSeed}
      disabled={state === "loading"}
      title={
        state === "success"
          ? "Data seeded!"
          : state === "error"
            ? "Seed failed"
            : "Seed sample data"
      }
      className={
        state === "success"
          ? "border-green-500 text-green-600"
          : state === "error"
            ? "border-red-500 text-red-600"
            : ""
      }
    >
      <Database className="size-4" />
      <span className="hidden sm:inline ml-1">
        {state === "loading"
          ? "Seeding…"
          : state === "success"
            ? "Seeded!"
            : state === "error"
              ? "Failed"
              : "Seed"}
      </span>
    </Button>
  );
}
