import { ResultPhase } from "@/components/game/impostor/ResultPhase";
import { AnimatePresence } from "framer-motion";

export default function TestPage() {
  return (
    <div className="pt-5">
      <AnimatePresence mode="wait">
        <ResultPhase key="result" />
      </AnimatePresence>
    </div>
  );
}
