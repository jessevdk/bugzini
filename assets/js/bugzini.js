var Service = {
    get: function(url, options) {
        var req = new XMLHttpRequest();

        req.open('GET', '/api' + url, true);

        req.onload = function(e) {
            var ret = req.responseText;
            var json = JSON.parse(ret);

            options.success(req, json);
        };

        req.onerror = function(e) {
            options.error(req);
        };

        req.onabort = function(e) {
            options.error(req);
        };

        req.send();
    }
};

var $ = function(elem) {
    this.elem = elem;
}

$.prototype.query = function(selector) {
    return this.elem.querySelector(selector);
}

$.prototype.queryAll = function(selector) {
    var nl = this.elem.querySelectorAll(selector);
    var ret = [];

    for(var i = nl.length; i--; ret.unshift(nl[i])) {}
    return ret;
}

var $$ = new $(document);

function _do_each(cb) {
    for (var i = 0; i < this.length; i++) {
        cb(this[i]);
    }
}

function _do_map(cb) {
    var ret = [];

    for (var i = 0; i < this.length; i++) {
        ret.push(cb(this[i]));
    }

    return ret;
}

function _do_filter(cb) {
    var ret = [];

    for (var i = 0; i < this.length; i++) {
        if (cb(this[i])) {
            ret.push(this[i]);
        }
    }

    return ret;
}

Array.prototype.each = _do_each;
Array.prototype.map = _do_map;
Array.prototype.filter = _do_filter;

NodeList.prototype.each = _do_each;

document.addEventListener('DOMContentLoaded', function() {
    document.removeEventListener('DOMContentLoaded', arguments.callee, false);

    app = new App();
    app.run();
}, false);

/* vi:ts=4:et */
