import type { Client } from "@heroiclabs/nakama-js";
import { useEffect, useState } from "react";
import * as nakamajs from "@heroiclabs/nakama-js";

interface PlayGroundProps {
  client: Client;
}

interface MatchState {
  presences: Record<string, nakamajs.Presence>;
  board: string[]; // flattened 3x3 board (length = 9)
  current_player: string; // userId of current turn
  started: boolean;
  winner: string;
}

// export default function TicTacToe() {
export default function PlayGround({ client }: PlayGroundProps) {
  const [board, setBoard] = useState<string[]>(Array(9).fill(""));
  //  const [isXTurn, setIsXTurn] = useState<boolean>(true);
  const [currentPlayer, setCurrentPlayer] = useState<string>("");
  const [socket, setSocket] = useState<nakamajs.Socket | null>(null);
  const [matchID, setMatchID] = useState<string>("");
  // const [userId, setUserId] = useState<string>("player123"); // example

  // const client = new nakamajs.Client("defaultkey", "localhost", "7350");
  // const player1: Player = { name: "Alice", symbol: "X" };
  // const player2: Player = { name: "Bob", symbol: "O" };

  // ------------------------
  // Connect to Nakama & Socket
  // ------------------------
  const connectSocket = async () => {
    const sessionToken = sessionStorage.getItem("sessionToken");
    if (!sessionToken) {
      console.error("❌ No session token found in storage.");
      return;
    }

    const sock = client.createSocket(false, false);
    sock.ondisconnect = (evt) => console.warn("Disconnected:", evt);

    try {
      await sock.connect({ token: sessionToken } as nakamajs.Session, true);
      console.info("✅ Connected to Nakama socket server.");
      setSocket(sock);

      // ------------------------
      // Listen for match updates
      // ------------------------
      sock.onmatchdata = (matchData) => {
        const stateStr = new TextDecoder().decode(matchData.data);
        const state: MatchState = JSON.parse(stateStr);
        console.log("match data ",state)
        setBoard(Array(9).fill(""));
        setCurrentPlayer(state.current_player);
        if (state.winner) {
          alert(
            state.winner === "draw" ? "It's a draw!" : `${state.winner} won!`
          );
        }
      };

      
      sock.onmatchpresence = (data) => {
        try {
            const member: nakamajs.Presence[] = data.joins;
            if (currentPlayer == ""){
              setCurrentPlayer(data.joins[0].user_id);
            }
            else{
              member.forEach((element,index) => {
                if (index>0){
                  alert(`${element.username} has joind the game`)
                }
              });
            }
            console.log("receive presensce update ",data)
          } catch (e) {
            console.error("Failed to parse match data:", e);
          }
        };

      return sock;
    } catch (err: any) {
      console.error("❌ Socket connection error:", err.message);
    }
  };

  // ------------------------
  // Find or Create Match
  // ------------------------
  const findOrCreateMatch = async () => {
    try {
      console.log("Start finding or creating match.....");
      if (!socket) return;
      const sessionToken = sessionStorage.getItem("sessionToken");
      if (!sessionToken) return;

      const session = { token: sessionToken } as nakamajs.Session;
      const payload = { user_id: "player123" };
      console.log("payload for client rpc ", payload);
      const response = await client.rpc(
        session,
        "find_or_create_match",
        payload
      );

      const matchData = response.payload as { match_id: string };
      const matchId = matchData.match_id;
      setMatchID(matchId.toString());
      const joinedMatch = await socket.joinMatch(matchId);
      console.log("✅ Joined match:", joinedMatch);
    } catch (error) {
      console.error(error);
    }
  };

  // ------------------------
  // Send Move to Server
  // ------------------------
  const sendMove = (index: number) => {
    console.log("cliecking ...")
    if (!socket || !matchID) return;
    try {
      // Check if it's your turn
      // if (currentPlayer !== userId) {
      //   console.warn("Not your turn");
      //   return;
      // }

      const payload = `${Math.floor(index / 3)},${index % 3}`; // row,col
      console.log("sending ....", payload," to match id ",matchID);
      socket.sendMatchState(matchID, 4, new TextEncoder().encode(payload))
        .then((res) => {
          console.log("move send successfully ", res);
        })
        .catch((err) => {
          console.log(err);
        }); // op_code = 0 for move
    } catch (error) {
      console.log(error);
    }
  };

  // ------------------------
  // Handle Cell Click
  // ------------------------
  const handleClick = (index: number) => {
    if (board[index]) return; // cell already filled
    console.log("Sending index ", index);
    sendMove(index);
  };

  // ------------------------
  // useEffect: Connect + Join Match
  // ------------------------
  useEffect(() => {
    const init = async () => {
      await connectSocket();
  
    };
    init();

    return () => {
      socket?.disconnect(true);
    };
  }, []);

  return (
    <div className="p-8 rounded-3xl bg-cyan-400 shadow-lg flex flex-col items-center">
      {/* Player Info */}
      <div className="flex gap-10 mb-8">
        Current turn: {currentPlayer === "player123" ? "You" : currentPlayer}
        {/* <p className="text-2xl">{player2.symbol}</p> */}
        {/* <div
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
        </div> */}
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
        onClick={() => findOrCreateMatch()}
        className="py-2 rounded-lg bg-green-600 text-white hover:bg-green-700 transition-all"
      >
        Create Room
      </button>
      <button
        onClick={() => findOrCreateMatch()}
        className="py-2 rounded-lg bg-green-600 text-white hover:bg-green-700 transition-all"
      >
        Join Room
      </button>
    </div>
  );
}
