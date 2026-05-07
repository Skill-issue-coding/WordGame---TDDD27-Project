import LobbyView from "@/components/lobby/LobbyView";
import { Loader2 } from "lucide-react";
import { Suspense } from "react";

async function LobbyParamUnwrapper({ params }: { params: Promise<{ lobbyCode: string }> }) {
  const { lobbyCode } = await params;

  return <LobbyView code={lobbyCode} />;
}

export default function Page({ params }: { params: Promise<{ lobbyCode: string }> }) {
  return (
    <Suspense
      fallback={
        <div className="flex flex-col items-center justify-center min-h-screen p-6">
          <Loader2 className="w-10 h-10 animate-spin text-game-purple mb-4" />
          <p className="font-display font-semibold text-muted-foreground">Hämtar rum...</p>
        </div>
      }>
      <LobbyParamUnwrapper params={params} />
    </Suspense>
  );
}
