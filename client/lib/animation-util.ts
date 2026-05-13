import type { MotionProps } from "framer-motion";

export type PopInMotion = Pick<MotionProps, "initial" | "animate" | "transition">;

export const popIn = (delay: number, popInStrength: number = 1): PopInMotion => ({
  initial: { opacity: 0, scale: 0.85, y: 14 },
  animate: { opacity: 1, scale: [0.85, popInStrength * 1.08, 1], y: [14, -4, 0] },
  transition: { delay, duration: 0.55, ease: "easeOut" },
});
