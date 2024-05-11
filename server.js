const { exec } = require("child_process");
const { log } = require("console");
const DHT = require("hyperdht");
const net = require("net");
const pump = require("pump");

const port = 3000;
const host = "localhost";

const rootPath = process.argv[2];
log(rootPath);
exec("main.exe", [rootPath], (error, stdout, stderr) => {});

const node = new DHT();

const server = node.createServer();

server.on("connection", function (socket) {
  pump(socket, net.connect(port, host), socket);
});

const keyPair = DHT.keyPair();
console.log(keyPair.publicKey.toString("hex"));

server.listen(keyPair);
