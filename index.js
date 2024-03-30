const net = require('net')


const port = 3000;
const host = '127.0.0.1';


const server = net.createServer();
server.listen(port, host, () => {
    console.log('TCP Server is running on port ' + port +'.');
});


server.on('connection', function(sock) {

    sock.on('data', function(data) {
        console.log(data)
        sock.write("reply\n");
    });

});