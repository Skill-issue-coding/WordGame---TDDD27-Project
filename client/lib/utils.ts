import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/**
 * The set of background colors available for player avatars.
 * Must be kept in sync with the `palette` slice in the Go util package.
 */
export const BACKGROUND_COLOR_PALETTE = ["#8b5cf6", "#ec4899", "#3b82f6", "#10b981", "#f59e0b", "#ef4444", "#06b6d4", "#a855f7"];

/**
 * Helper function to check if a string is empty or contains white spaces
 */
export function isStringEmptyOrOnlySpaces(str: string) {
  return str === null || str.match(/^ *$/) !== null;
}
