
import './App.css'
import PlayGround from './pages/Playground'
import {Client} from "@heroiclabs/nakama-js";

function App() {
  
  var useSSL = false; // Enable if server is run with an SSL certificate.
  var client = new Client("defaultkey", "127.0.0.1", "7350", useSSL);
  client.authenticateCustom("someuniqueid", true,"soumyaRanjan",).then((session) => {
    console.log("Authenticated custom id. Session token: " + session.token);
  }).catch((error) => {
    console.log("Error authenticating custom id: " + error.message);
  });

  return (
    <>
      <h1 className="text-3xl font-bold mb-6 text-gray-800">Tic Tac Toe</h1>
    <PlayGround/> 
    </>
  )
}

export default App
