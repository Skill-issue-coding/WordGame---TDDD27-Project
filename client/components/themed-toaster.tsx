"use client";

import { Toaster } from "sonner";
import { useTheme } from "next-themes";

export default function ThemedToaster() {
  const { resolvedTheme } = useTheme();

  return <Toaster position="top-right" theme={resolvedTheme === "dark" ? "dark" : "light"} />;
}
