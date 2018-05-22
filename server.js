'use strict';

const http = require('http');
const os = require('os');
const cluster = require('cluster');

const port = process.env.PORT || 3000;
const cpus = os.cpus();

if (cluster.isMaster) {
  console.log('Node version:', process.version);
  console.log('CPU: %d x %s', cpus.length, cpus[0].model);

  if (!process.env.CLUSTER) {
    runServer();
  } else {
    console.log(`Master ${process.pid} is running`);

    // Fork workers.
    for (let i = 0; i < cpus.length; i++) {
      cluster.fork();
    }
  }
  console.log('Listening on port %s', port);
} else {
  console.log(`Worker ${process.pid} started`);

  runServer();
}

function runServer() {
  let server = http.createServer((request, response) => {
    const { headers, method, url } = request;
    let body = [];
    request.on('error', (err) => {
      console.error(err);
    }).on('data', (chunk) => {
      body.push(chunk);
    }).on('end', () => {
      body = JSON.parse(Buffer.concat(body).toString());

      response.statusCode = 200;
      response.setHeader('Content-Type', 'application/json');
      response.end(JSON.stringify(body))
    });
  })

  server.listen(port);
}