const initCluster = require('com_github_kalbasit_bazel/server/cluster');
const server = require('com_github_kalbasit_bazel/server/server');

const clusterOpts = {
  count: parseInt(process.env.NUM_WORKERS, 10),
};

initCluster((cfg) => {
  console.log(`[${cfg.name}] listening on :8080`);
}, () => {
  server.init();
}, clusterOpts);
