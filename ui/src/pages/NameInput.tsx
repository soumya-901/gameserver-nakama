import {type Dispatch, type SetStateAction } from "react";

interface InputNameProps {
    name: string;
    setName: Dispatch<SetStateAction<string>>
    CustomAuthentication: (username: string) => void;
}

const NameInput: React.FC<InputNameProps> = ({ name,setName,CustomAuthentication }) => {


  return (

    <>
        <h2 className="text-xl font-semibold text-gray-800 mb-4 text-center">
          Enter Your Name
        </h2>
        <input
          type="text"
          name="usernaem"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Your name..."
          className="w-full p-[5px] bg-transparent border mb-4 focus:outline-none focus:ring-2 focus:ring-blue-400"
          style={{ borderRadius: "8px", color: "black"}}
        />

        <div className="text-center mb-4 text-gray-700">
              <p className="text-lg font-medium">
              👋 Hi, { name && <span className="text-blue-600">{name} !</span> }
            </p>
        </div>

        <button
          onClick={() => CustomAuthentication(name)}
          className="py-2 rounded-lg bg-green-600 text-white hover:bg-green-700 transition-all"
        >
          Play
        </button>
            </>
    //   </Modal>
    // </div>
  );
}
export default NameInput;
