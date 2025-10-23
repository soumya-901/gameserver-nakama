import type { Client } from "@heroiclabs/nakama-js";
import { useEffect, useRef, useState } from "react";
import * as nakamajs from "@heroiclabs/nakama-js";
import style from "./styles_pages.module.css";
import { toast } from "react-toastify";

interface PlayGroundProps {
  client: Client;
}

interface MatchState {
  match_id: string;
  presences: Record<string, nakamajs.Presence>;
  board: string[][]; // flattened 3x3 board (length = 9)
  current_player: string; // userId of current turn
  current_symbol: string; // userId of current turn
  started: boolean;
  winner: string;
}

// export default function TicTacToe() {
export default function PlayGround({ client }: PlayGroundProps) {
  const [board, setBoard] = useState<string[][]>(
    Array(3).fill(Array(3).fill(""))
  );
  //  const [isXTurn, setIsXTurn] = useState<boolean>(true);

  const [currentPlayer, setCurrentPlayer] = useState<string>("");
  const [currentSymbol, setCurrentSymbol] = useState<string>("");
  const [socket, setSocket] = useState<nakamajs.Socket | null>(null);
  const [matchID, setMatchID] = useState<string>("");
  const [userId, setUserId] = useState<string>(""); // example
  const current_player_ref = useRef<string>("");
  const user_id_ref = useRef<string>("");

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
        console.log("match data ", state);
        setCurrentSymbol(state.current_symbol);
        setBoard(state.board);
        setCurrentPlayer(state.current_player);
        current_player_ref.current = state.current_player;
        if (state.winner) {
          if (state.winner === "draw") {
            toast.info("It's a draw!");
          } else if (user_id_ref.current == state.winner) {
            toast.success("✨ You Won ✨");
          } else {
            toast.info("You Lost 🥲");
          }
          current_player_ref.current = "";
          user_id_ref.current = "";
          setCurrentPlayer("");
          setUserId("");
          setMatchID("");
        }
      };

      sock.onmatchpresence = (data) => {
        try {
          const member: nakamajs.Presence[] = data.joins;
          console.log(
            "current player turn ",
            current_player_ref.current,
            " player id ",
            user_id_ref.current
          );
          if (user_id_ref.current === "") {
            setUserId(data.joins[0].user_id);
            user_id_ref.current = data.joins[0].user_id;
          } else {
            member.forEach((element) => {
              toast.info(`${element.username} has joind the game`);
            });
          }
          console.log(
            "inside id updation ",
            user_id_ref.current,
            " current user ",
            current_player_ref.current
          );
          console.log("receive presensce update ", data);
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
    setBoard(Array(3).fill(Array(3).fill("")));
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
      console.log("JOINing to the match ", matchId);
      const joinedMatch = await socket.joinMatch(matchId);
      console.log("✅ Joined match:", joinedMatch);
    } catch (error) {
      console.error(error);
    }
  };

  // ------------------------
  // Send Move to Server
  // ------------------------
  const sendMove = (row: number, col: number) => {
    if (!socket || !matchID) return;
    console.log("cliecking ...");
    try {
      const payload = `${row},${col}`; // row,col
      console.log("sending ....", payload, " to match id ", matchID);
      socket
        .sendMatchState(matchID, 4, new TextEncoder().encode(payload))
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
  const handleClick = (row: number, column: number) => {
    if (board[row][column]) return; // cell already filled
    // Check if it's your turn
    if (currentPlayer !== userId) {
      console.log("current player ", currentPlayer, " user id-", userId);
      toast.warn("Not your Turn");
      return;
    }
    console.log("Sending index ", row, column);
    setBoard((prevBoard) =>
      prevBoard.map((r, rIdx) =>
        rIdx === row
          ? r.map((c, cIdx) =>
              cIdx === column && c === "" ? currentSymbol : c
            )
          : r
      )
    );
    // setCurrentPlayer(prev => (prev === "X" ? "O" : "X"));
    sendMove(row, column);
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
      <div style={{ marginBottom: "1rem", fontSize: "20px" }}>
        {matchID == "" ? (
          ""
        ) : currentPlayer == "" ? (
          <div>
            Waitting For Opponent <div className={style.loader}></div>
          </div>
        ) : currentPlayer === userId ? (
          "Your Turn"
        ) : (
          "Opponent turn"
        )}
      </div>

      {/* Game Board */}
      <div className="grid grid-cols-3 gap-4">
        {board.map((row, rowIndex) =>
          row.map((value, colIndex) => (
            <div
              key={`${rowIndex}-${colIndex}`}
              onClick={() => handleClick(rowIndex, colIndex)}
              className="text-3xl font-bold flex justify-center items-center transition-all rounded-none bg-cyan-500 hover:bg-cyan-400 active:bg-cyan-600"
              style={{
                width: "100px",
                height: "100px",
                cursor: "pointer",
                borderRight:
                  colIndex === 2 ? "3px solid #22d3ee" : "3px solid grey",
                borderBottom:
                  rowIndex === 2 ? "3px solid #22d3ee" : "3px solid grey",
              }}
            >
              <span className="scale-210">{value}</span>
            </div>
          ))
        )}
      </div>

      {!matchID && (
        <button
          onClick={() => findOrCreateMatch()}
          className="rounded-lg bg-green-600 text-white hover:bg-green-700 transition-all"
          style={{ marginTop: "1rem" }}
        >
          Join Game
        </button>
      )}
    </div>
  );
}
