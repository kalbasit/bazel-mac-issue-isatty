var lib = require('com_github_kalbasit_bazel/server/lib');

function init() {
  var server = http.createServer(function(req, res) {
    var url = lib.parseURL(req.url);
		process.send(url);
  });
  server.listen(8080);
}

module.exports = {
  init,
};
