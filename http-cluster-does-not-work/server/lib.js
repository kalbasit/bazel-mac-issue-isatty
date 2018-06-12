function cleanQuery(url) {
	if (!url) {
		return '';
	}
	let u = ('' + url).trim();
	const idx = u.indexOf('?');
	if (idx > -1) {
		return u.substring(idx + 1);
	}
	return u;
}

function parseQuery(url) {
	const query = {};
	cleanQuery(url).split('&').forEach(function (pa) {
		const [key, val] = pa.split('=');
		query[decodeURIComponent(key || '').trim()] = decodeURIComponent(val || '').trim();
	});
	return query;
}

function parseURL(url) {
	const idx = url.indexOf('?');
	if (idx === -1) {
		return {
			path: url,
			params: {},
		};
	}
	return {
		path: url.substring(0, idx),
		params: parseQuery(url),
	};
}

module.exports = {
	parseURL,
	parseQuery,
	cleanQuery,
};
