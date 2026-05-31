import { ReactNode } from "react";
import { motion } from "framer-motion";

interface PhaseTransitionProps {
  phaseKey: string;
  children: ReactNode;
}

const PhaseTransition = ({ phaseKey, children }: PhaseTransitionProps) => (
  <motion.div
    key={phaseKey}
    initial={{ opacity: 0, y: 24, scale: 0.96 }}
    animate={{ opacity: 1, y: 0, scale: 1 }}
    exit={{ opacity: 0, y: -24, scale: 0.96 }}
    transition={{ duration: 0.35, ease: [0.22, 1, 0.36, 1] }}
    className="flex flex-col items-center w-full">
    {children}
  </motion.div>
);

export default PhaseTransition;
