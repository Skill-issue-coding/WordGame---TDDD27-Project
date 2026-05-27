import { VotePhase } from "@/components/game/impostor/VotePhase";
import { AnimatePresence } from "framer-motion";

export default function TestPage() {
  return (
    <div className="pt-5">
      <AnimatePresence mode="wait">
        <VotePhase key="vote" />
      </AnimatePresence>
    </div>
  );
}
