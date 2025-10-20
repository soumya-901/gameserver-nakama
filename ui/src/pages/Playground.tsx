import { useState } from "react";

interface Player {
  name: string;
  symbol: "X" | "O";
}

export default function PlayGround() {

  const [board, setBoard] = useState<string[]>(Array(9).fill(""));
  const [isXTurn, setIsXTurn] = useState<boolean>(true);

  const player1: Player = { name: "Alice", symbol: "X" };
  const player2: Player = { name: "Bob", symbol: "O" };

  const handleClick = (index: number) => {
    if (board[index]) return; // ignore already filled cells

    const newBoard = [...board];
    newBoard[index] = isXTurn ? player1.symbol : player2.symbol;
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
              ...((index+1)%3 ==0 ? { borderBottom:"3px solid grey",borderRight:"3px solid #22d3ee" } : {borderBottom:"3px solid grey", borderRight:"3px solid grey"}),
              ...(index>=6 ? { borderBottom:"3px solid #22d3ee" } : {}),
            }}
                
          >
            <span className="scale-210">

            {value}
            </span>
          </div>
        ))}
      </div>
      </div>
  );
}