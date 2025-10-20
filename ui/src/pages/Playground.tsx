import type { Client, Session, Socket } from "@heroiclabs/nakama-js";
import { useEffect, useState } from "react";

interface Player {
  name: string;
  symbol: "X" | "O";
}

interface PlayGroundProps {
  client: Client;
}

export default function PlayGround({ client }: PlayGroundProps) {
  const [board, setBoard] = useState<string[]>(Array(9).fill(""));
  const [isXTurn, setIsXTurn] = useState<boolean>(true);
  const [socketCon, setSocketCon] = useState<Socket | null>(null);
  const [matchID , setMatchID ] = useState<string>("")

  const player1: Player = { name: "Alice", symbol: "X" };
  const player2: Player = { name: "Bob", symbol: "O" };

  useEffect(() => {
    const connectSocket = async () => {
      const sessionToken = sessionStorage.getItem("sessionToken"); // use sessionStorage for consistency
      if (!sessionToken) {
        console.error("❌ No session token found in storage.");
        return;
      }

      const trace = false;
      const socket = client.createSocket(false, trace);

      socket.ondisconnect = (evt) => {
        console.info("⚠️ Disconnected:", evt);
      };

      const session: Session = {
        token: sessionToken,
        refresh_token: sessionStorage.getItem("refreshToken") || "",
        created_at:
          Number(sessionStorage.getItem("createdAt")) || Date.now() / 1000,
      } as Session;

      try {
        console.log("🧠 Connecting to Nakama socket...");
        await socket.connect(session, true);
        console.info("✅ Connected to Nakama socket server.");
        setSocketCon(socket);

        socket.onmatchpresence = (data) => {
          try {
            console.log("receive presensce update ",data)
          } catch (e) {
            console.error("Failed to parse match data:", e);
          }
        };

        socket.onmatchdata = (data) => {
          try {
            console.log("on choosing x or 0 ",data)
          } catch (err) {
            console.error("Error decoding match data:", err);
          }
        };

        // // 🟩 Create a new match if needed
        // const match = await socket.createMatch("tictactoe"); // must match your Go handler name
        // console.log("✅ Created match:", match);

        // // 🟨 Join that match
        // const joined = await socket.joinMatch(match.match_id);
        // console.log("✅ Joined match:", joined);
      } catch (error: any) {
        console.error("❌ Socket or match error:", error.message);
      }
    };

    connectSocket();

    // optional cleanup
    return () => {
      socketCon?.disconnect(true);
    };
  }, [client]);

  // 🧩 Helper functions (if you want to call them manually later)
  async function createMatch() {
    try {
      const match = await socketCon?.createMatch("tictactoe");
      console.log("✅ Created match:", match);
      return match?.match_id;
    } catch (err: any) {
      console.error("❌ Failed to create match:", err.message);
    }
  }

  async function joinMatch(matchId: string) {
    try {
      console.log("🔗 Joining match...");
      const sessionToken = sessionStorage.getItem("sessionToken");

      // prefer explicit matchId, otherwise fetch top match from server
      let targetMatchId: string = matchId;
      if (!targetMatchId) {
        const topMatch = await client.listMatches({
          token: sessionToken,
        } as Session);
        const firstMatch =
          topMatch.matches && topMatch.matches.length > 0
            ? topMatch.matches[0]
            : undefined;
        if (!firstMatch) {
          console.error("❌ No matches available to join.");
          return;
        }
        const fmId = firstMatch.match_id;
        if (fmId === undefined || fmId === null) {
          console.error("❌ Top match has no match_id.");
          return;
        }
        targetMatchId = fmId.toString();
        setMatchID(targetMatchId)
      }

      if (!socketCon) {
        console.error("❌ Socket not connected.");
        return;
      }

      const match = await socketCon.joinMatch(targetMatchId);
      console.log("✅ Joined match:", match);
    } catch (err: any) {
      console.error("Error joining match:", err.message);
    }
  }

  function sendSymbolChoice(symbol:string) {
  if (!socketCon) {
    console.error("Socket not connected!");
    return;
  }

  const payload = { symbol };
  const encoded = new TextEncoder().encode(JSON.stringify(payload));
  socketCon.sendMatchState(matchID, 2, encoded); // opcode 2 = symbol select
  console.log(`Sent symbol choice: ${symbol}`);
}


  const handleClick = (index: number) => {
    if (board[index]) return; // ignore already filled cells

    const newBoard = [...board];
    newBoard[index] = isXTurn ? player1.symbol : player2.symbol;
    sendSymbolChoice(newBoard[index])
    setBoard(newBoard);
    setIsXTurn(!isXTurn);
  };

  return (
    <div className="p-8 rounded-3xl bg-cyan-400 shadow-lg flex flex-col items-center">
      {/* Player Info */}
      <div className="flex gap-10 mb-8">
        <div
          className={`p-4 rounded-2xl shadow ${
            isXTurn ? "bg-blue-200" : "bg-white"
          }`}
        >
          <p className="font-semibold text-lg">{player1.name}</p>
          <p className="text-2xl">{player1.symbol}</p>
        </div>

        <div
          className={`p-4 rounded-2xl shadow ${
            !isXTurn ? "bg-blue-200" : "bg-white"
          }`}
        >
          <p className="font-semibold text-lg">{player2.name}</p>
          <p className="text-2xl">{player2.symbol}</p>
        </div>
      </div>

      {/* Game Board */}
      <div className="grid grid-cols-3 gap-4">
        {board.map((value, index) => (
          <div
            key={index}
            onClick={() => handleClick(index)}
            className="text-3xl font-bold flex justify-center  transition-all rounded-none bg-cyan-500 hover:bg-cyan-400 active:bg-cyan-600"
            style={{
              width: "100px",
              height: "100px",
              display: "flex",
              cursor: "pointer",
              alignItems: "center",
              //   backgroundColor: 'var(--color-cyan-500)',
              ...((index + 1) % 3 == 0
                ? {
                    borderBottom: "3px solid grey",
                    borderRight: "3px solid #22d3ee",
                  }
                : {
                    borderBottom: "3px solid grey",
                    borderRight: "3px solid grey",
                  }),
              ...(index >= 6 ? { borderBottom: "3px solid #22d3ee" } : {}),
            }}
          >
            <span className="scale-210">{value}</span>
          </div>
        ))}
          </div>
        <button
          onClick={() => createMatch()}
          className="py-2 rounded-lg bg-green-600 text-white hover:bg-green-700 transition-all"
        >
          Create Room
        </button>
        <button
          onClick={() => joinMatch("")}
          className="py-2 rounded-lg bg-green-600 text-white hover:bg-green-700 transition-all"
        >
          Join Room
        </button>
    </div>
  );
}
