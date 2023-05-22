import { exec } from "https://deno.land/x/exec/mod.ts";
import { load } from "https://deno.land/std/dotenv/mod.ts";
import { delay } from "https://deno.land/std/async/mod.ts";

const configDir = "./ovpn";
const servers = [
  "node-dk-03.protonvpn.net.tcp",
  "node-fr-14.protonvpn.net.tcp",
  "node-fr-15.protonvpn.net.tcp",
];

const env = await load();
const username = env["USERNAME"];
const password = env["PASSWORD"];

// Function to handle Ctrl+C
function handleSignal() {
  console.log("\nReceived Ctrl+C. Disconnecting from all VPN connections...");
  exec("pkill openvpn");
  Deno.exit(0);
}

// IP address before connecting to VPN
const response = await exec("curl -s https://ifconfig.co");
console.log(`Your IP address before connecting to VPN is: ${response.output}`);

// Register the Ctrl+C signal handler
Deno.addSignalListener("SIGINT", handleSignal);

while (true) {
  for (const server of servers) {
    const configFile = `${configDir}/${server}.ovpn`;
    console.log(`Connecting to ${configFile}`);

    // Run OpenVPN with credentials provided
    // TODO

    // Wait for the VPN connection to be established
    await delay(2000);

    // Retrieve the VPN IP address
    const vpnIPResponse = await exec("curl -s https://ifconfig.co");
    console.log(`Your VPN IP is: ${vpnIPResponse.output}`);

    // Wait for 60 seconds before disconnecting
    await delay(60000);

    console.log(`Disconnecting from ${configFile}`);
    openvpnProcess.close();
    await delay(2000);
  }
}
