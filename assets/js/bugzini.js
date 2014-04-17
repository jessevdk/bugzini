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

Object.prototype.each = function (cb) {
    for (var i = 0; i < this.length; i++) {
        cb(this[i]);
    }
}

Object.prototype.map = function(cb) {
    var ret = [];

    for (var i = 0; i < this.length; i++) {
        ret.push(cb(this[i]));
    }

    return ret;
}

Object.prototype.filter = function(cb) {
    var ret = [];

    for (var i = 0; i < this.length; i++) {
        if (cb(this[i])) {
            ret.push(this[i]);
        }
    }

    return ret;
}

Object.prototype.$ = function() {
    return new $(this);
}

document.addEventListener('DOMContentLoaded', function() {
    document.removeEventListener('DOMContentLoaded', arguments.callee, false);

    var app = new App();
    app.run();
}, false);

/* vi:ts=4:et */
