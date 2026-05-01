"use client";

import { useState } from "react";
import { Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cleanData } from "@/lib/api";

type State = "idle" | "loading" | "success" | "error";

export function CleanButton() {
  const [state, setState] = useState<State>("idle");

  async function handleClean() {
    if (state === "loading") return;
    setState("loading");
    const result = await cleanData();
    setState(result === null ? "error" : "success");
    setTimeout(() => setState("idle"), 3000);
  }

  return (
    <Button
      variant="outline"
      size="sm"
      onClick={handleClean}
      disabled={state === "loading"}
      title={
        state === "success"
          ? "Data cleaned!"
          : state === "error"
            ? "Clean failed"
            : "Clean all catalog data"
      }
      className={
        state === "success"
          ? "border-green-500 text-green-600"
          : state === "error"
            ? "border-red-500 text-red-600"
            : ""
      }
    >
      <Trash2 className="size-4" />
      <span className="hidden sm:inline ml-1">
        {state === "loading"
          ? "Cleaning…"
          : state === "success"
            ? "Cleaned!"
            : state === "error"
              ? "Failed"
              : "Clean"}
      </span>
    </Button>
  );
}
