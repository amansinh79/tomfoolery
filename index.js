const { spawn } = require("child_process");
const DHT = require("hyperdht");
const net = require("net");
const pump = require("pump");

const node = new DHT();

const server = net.createServer();
server.listen(4000, () => {
  console.log("TCP Server is running on port " + 4000 + ".");
});

server.on("connection", function (s) {
  pump(s, node.connect(Buffer.from(process.argv[2], "hex")), s);
});

spawn("main.exe", {
  stdio: "inherit",
});
