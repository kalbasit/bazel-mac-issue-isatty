var cluster = require('cluster'),
    os = require('os');

function config(cfg) {
  if (!cfg) {
    cfg = {};
  }
  cfg.name = 'name' in cfg ? cfg.name : `server-${(new Date).getTime()}`;
  cfg.count = +cfg.count > 0 ? cfg.count : os.cpus().length;
  return cfg;
}

function noop() {}

function restartProcess(cfg, worker) {
  worker.kill();
  // Log deaths!
  console.log(`[${cfg.name}] Worker ${worker.process.pid} died`);
  // If autoRestart is true, spin up another to replace it
  process.nextTick(function () {
    forkProcess(cfg);
  });
}

function forkProcess(cfg) {
  console.log(`[${cfg.name}] starting worker thread`);
  const worker = cluster.fork();
  let idx = 0;
  worker.on('message', function(msg) {
		console.log(`received message ${msg}`);
  });
  worker.on('exit', function() {
    restartProcess(cfg, worker);
  });
}

module.exports = function initCluster(masterCB, clusterCB, cfg) {
  cfg = config(cfg);
  if (cluster.isMaster) {
    console.log(`[${cfg.name}] Starting master`);
    (masterCB || noop)(cfg);

    console.log(`[${cfg.name}] Cluster options: ${ JSON.stringify(cfg, null, 4) }`);
    for (let i = 0; i < cfg.count; i += 1) {
      forkProcess(cfg);
    }
    cluster.on('death', function (worker) {
      restartProcess(cfg, worker);
    });
  } else {
    console.log(`[${cfg.name}] Worker ${process.pid} started`);
    (clusterCB || noop)(cfg);
  }
};
