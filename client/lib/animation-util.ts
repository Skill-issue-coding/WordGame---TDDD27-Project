import type { MotionProps } from "framer-motion";

export type PopInMotion = Pick<MotionProps, "initial" | "animate" | "transition">;

export type PopInOptions = {
  delay?: number;
  strength?: number;
  y?: number;
  duration?: number;
  ease?: NonNullable<MotionProps["transition"]>["ease"];
};

export type SnapInOptions = PopInOptions & {
  x?: number;
  rotate?: number;
};

const resolvePopInOptions = (delayOrOptions?: number | PopInOptions, strength?: number): Required<PopInOptions> => {
  if (typeof delayOrOptions === "number") {
    return {
      delay: delayOrOptions,
      strength: strength ?? 1,
      y: 14,
      duration: 0.55,
      ease: "easeOut",
    };
  }

  return {
    delay: delayOrOptions?.delay ?? 0,
    strength: delayOrOptions?.strength ?? 1,
    y: delayOrOptions?.y ?? 14,
    duration: delayOrOptions?.duration ?? 0.55,
    ease: delayOrOptions?.ease ?? "easeOut",
  };
};

export const popIn = (delayOrOptions?: number | PopInOptions, strength?: number): PopInMotion => {
  const { delay, strength: resolvedStrength, y, duration, ease } = resolvePopInOptions(delayOrOptions, strength);

  return {
    initial: { opacity: 0, scale: 0.85, y },
    animate: { opacity: 1, scale: [0.85, resolvedStrength * 1.08, 1], y: [y, -4, 0] },
    transition: { delay, duration, ease },
  };
};

export const snapIn = (options: SnapInOptions = {}): PopInMotion => {
  const { delay, strength, y, duration, ease } = resolvePopInOptions(options);
  const x = options.x ?? 0;
  const rotate = options.rotate ?? -2;

  return {
    initial: { opacity: 0, scale: 0.9, y, x, rotate },
    animate: {
      opacity: 1,
      scale: [0.9, strength * 1.04, 1],
      y: [y, -2, 0],
      x: [x, 0, 0],
      rotate: [rotate, rotate * -0.3, 0],
    },
    transition: { delay, duration: duration * 0.75, ease },
  };
};
