
import { useEffect, useState } from 'react';
import './App.css'
import PlayGround from './pages/Playground'
import {Client} from "@heroiclabs/nakama-js";
import NameInput from './pages/NameInput';

function App() {

  const [accoutDetails, setAccountDetails] = useState<any>(null);
  const [nakamaClient, setNakamaClient] = useState<Client>({} as Client);
  const [name, setName] = useState("");
  useEffect(() => { 
    // fetch from local storage
    var useSSL = false; // Enable if server is run with an SSL certificate.
    var client = new Client("defaultkey", "127.0.0.1", "7350", useSSL);
    const trace = false;
    const socket = client.createSocket(useSSL, trace);
    socket.ondisconnect = (evt) => {
        console.info("Disconnected", evt);
    };
    setNakamaClient(client);
    const sessionToken = localStorage.getItem("sessionToken");
    if (sessionToken) {
      // client.getAccount expects a Session object; create a minimal object with the token
      client.getAccount({ token: sessionToken } as any).then((accountDetails) => {
        console.log("Account details:", accountDetails);
        setAccountDetails(accountDetails);
      }).catch((error) => {
        console.log("Error fetching account details: " + error.message);
      });
      socket.connect({token:sessionToken}as any,true).then(() => {
          console.info("Connected to nakama socket server.");
      }).catch((error) => {
          console.error("Failed to connect to Nakama socket server:", error);
      });
    }

  }, []);

  function CustomAuthentication(username: string) {
        nakamaClient.authenticateCustom(`${username}someuniqueid`, true, username).then((session) => {
          console.log("Authenticated custom id. Session token: " + session.token);
          localStorage.setItem("sessionToken", session.token);
          // Now you can fetch account details
          nakamaClient.getAccount(session).then((accountDetails) => {
            console.log("Account details:", accountDetails);
            setAccountDetails(accountDetails);
          }).catch((error) => {
            console.log("Error fetching account details: " + error.message);
          });
        }).catch((error) => {
          console.log("Error authenticating custom id: " + error.message);
        });
  }

  return (
    <>
      <h1 className="text-3xl font-bold mb-6 text-gray-800">Tic Tac Toe</h1>
      {accoutDetails ? (
        <PlayGround  />) 
        : (
   
            <NameInput name={name} setName={setName} CustomAuthentication={CustomAuthentication} />

        )
      }
    </>
  )
}

export default App
