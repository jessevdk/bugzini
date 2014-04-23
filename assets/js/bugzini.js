var escapeDiv = document.createElement('div');
var escapeElement = document.createTextNode('');
escapeDiv.appendChild(escapeElement);

function html_escape(s) {
    escapeElement.data = s;
    return escapeDiv.innerHTML;
}

var Service = {
    ajax: function(method, url, options) {
        var req = new XMLHttpRequest();

        req.open(method, '/api' + url, true);
        var fd = null;

        if (options.data) {
            fd = new FormData();

            for (var k in options.data) {
                fd.append(k, options.data[k]);
            }
        }

        req.onload = function(e) {
            if (req.status != 200) {
                if (options.error) {
                    options.error(req);
                }
            } else {
                var ret = req.responseText;
                var json = null;

                if (ret) {
                    json = JSON.parse(ret);
                }

                options.success(req, json);
            }
        };

        req.onerror = function(e) {
            if (options.error) {
                options.error(req);
            }
        };

        req.onabort = function(e) {
            if (options.error) {
                options.error(req);
            }
        };

        req.send(fd);
    },

    get: function(url, options) {
        Service.ajax('GET', url, options);
    },

    post: function(url, options) {
        Service.ajax('POST', url, options);
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
