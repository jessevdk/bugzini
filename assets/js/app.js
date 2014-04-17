var App = function() {
    this.init();
}

App.prototype.init = function() {
    var searches = $$.queryAll('.search');

    this.searches = {};

    searches.each((function (e) {
        this.searches[e.id] = new Search(e);
    }).bind(this));

    this.filters = $$.query('#filters');
    this._filters = [];
    this._search_filter = '';
    this._filter_map = {};

    this.searches['search-filters'].on_update = this.on_search_filters.bind(this);

    var items = $$.query('#sidebar_items')

    items.addEventListener('DOMSubtreeModified', function (e) {
        if (items.scrollHeight > items.clientHeight) {
            items.classList.remove('hide-scrollbar')
        } else {
            items.classList.add('hide-scrollbar')
        }
    });

    var all_starred = $$.query('#all-starred-filter');

    all_starred.addEventListener('click', (function() {
        this._on_filter_click.bind(this)(all_starred);
    }).bind(this));

    this.db = new DB();
    this.db.on_filters_updated = this.on_filters_updated.bind(this);

    this._active_filter_elem = all_starred;
    this._active_filter = null;

    this.db.loaded = (function() {
        this._update_bugs_list();
    }).bind(this);
}

App.prototype.on_search_filters = function(search) {
    this._search_filter = search.value().toLowerCase();
    this._update_filters();
}

App.prototype.on_filters_updated = function() {
    this.db.filters().all((function(filters) {
        this._filters = filters;
        this._filter_map = {};

        filters.each((function(f) {
            this._filter_map[f.id] = f;
        }).bind(this));

        this._update_filters();
    }).bind(this));
}

App.prototype._tokenize_query = function(query) {
    var tokens = [];
    var i = 0;

    var schar = '!()" '

    while (i < query.length) {
        switch (query[i]) {
        case '!':
            tokens.push({type: '!'});
            break;
        case '(':
            tokens.push({type: '('});
            break;
        case ')':
            tokens.push({type: ')'});
            break;
        case '"':
            i++;
            var s = i;

            while (i < query.length && query[i] != '"') {
                i++;
            }

            tokens.push({type: 's', value: query.slice(s, i)});
            break;
        case ' ':
            break;
        default:
            var s = i;

            while (i < query.length && schar.indexOf(query[i]) == -1) {
                i++;
            }

            tokens.push({type: 's', value: query.slice(s, i)});
            i--;
            break;
        }

        i++;
    }

    return tokens;
}

App.prototype._parse_tokens = function(tokens) {
    var root = {
        either: [],
        all: []
    };

    var stack = [root];
    var i = 0;

    var ungroup = function() {
        var p = stack.shift();

        if (p.all.length == 0) {
            stack[0].either = stack[0].either.concat(p.either);
        } else {
            stack[0].either.push(p);
        }
    }

    var parse = function() {
        var tok = tokens[i];

        switch (tok.type) {
        case '(':
            i++;
            stack.unshift({either: [], all: []});

            var l = stack.length;

            while (i < tokens.length && stack.length >= l) {
                parse();
            }

            return;
        case ')':
            ungroup();
            break;
        case '!':
            i++;

            var neg = {either: [], all: []};

            stack.unshift(neg);
            parse();
            stack.shift();

            stack[0].all.push(neg);
            return;
        case 's':
            stack[0].either.push(tok.value)
            break;
        }

        i++;
    }

    while (i < tokens.length) {
        parse();
    }

    while (stack.length > 1) {
        ungroup();
    }

    return root;
}

App.prototype._flat_queries = function(query) {
    
}

App.prototype._parse_query = function(query) {
    var tokens = this._tokenize_query(query);
    var q = this._parse_tokens(tokens);

    
    console.log(JSON.stringify(q));
}

App.prototype._update_bugs_list_with_filters = function(filters) {
    var q = '!(' + filters.map(function(e) { return e.query; }).join(' ') + ')';
    var s = this.searches['search-bugs'].value();

    if (s) {
        q += ' (' + s + ')';
    }

    this._parse_query(q);
    console.log(q);
}

App.prototype._update_bugs_list = function() {
    // Get all bugs according to the current filter + search terms
    var filter = '';

    if (this._active_filter == null) {
        // Filter all starred
        this.db.filters().index('starred').only(1).all(this._update_bugs_list_with_filters.bind(this));
    } else {
        this._update_bugs_list_with_filters([this._active_filter]);
    }
}

App.prototype._on_filter_click = function(elem, filter) {
    if (this._active_filter != filter) {
        if (this._active_filter_elem) {
            this._active_filter_elem.classList.remove('selected');
        }

        this._active_filter_elem = elem;
        this._active_filter = filter;

        elem.classList.add('selected');
    }

    this._update_bugs_list();
}

App.prototype._filter_before = function(filter) {
    var dbf = this.db.filters();

    var ret = {
        oindex: -1,
        index: -1
    };

    for (var i = 0; i < this._filters.length; i++) {
        if (filter == this._filters[i]) {
            ret.index = i;

            if (ret.oindex != -1) {
                break;
            }
        } else if (ret.oindex == -1 && dbf.sort(filter, this._filters[i]) <= 0) {
            ret.oindex = i;

            if (ret.index != -1) {
                break;
            }
        }
    }

    return ret;
}

App.prototype._on_filter_star_click = function(elem, star, filter) {
    filter.starred = filter.starred ? 0 : 1;

    this.db.filters().put(filter, (function() {
        var starred = filter.starred;

        if (starred) {
            star.classList.remove('non-starred');
            star.classList.add('starred');
        } else {
            star.classList.remove('starred');
            star.classList.add('non-starred');
        }

        // Slide out/in newly sorted
        var parent = elem.parentNode;
        parent.removeChild(elem);

        var before = this._filter_before(filter);

        this._filters.splice(before.index, 1);
        this._filters.splice(before.oindex, 0, filter);

        this._update_filters();
    }).bind(this));
}

App.prototype._update_filters = function() {
    var html = '';
    var had_starred = false;
    var is_products = false;

    this._filters.each((function(filter) {
        if (this._search_filter && filter.name_case.indexOf(this._search_filter) == -1) {
            return;
        }

        if (!filter.starred && had_starred) {
            html += '<li class="separator"></li>';
            had_starred = false;
        }

        if (filter.starred) {
            had_starred = true;
        }

        if (!is_products && filter.is_product && !filter.starred) {
            is_products = true;
            html += '<li class="header">Products</li>';
        }

        if (filter.starred) {
            cls = 'starred';
        } else {
            cls = 'non-starred';
        }

        var licls = 'item';

        if (filter == this._active_filter) {
            licls += ' selected'
        }

        html += '<li class="' + licls + '" data-id="' + filter.id + '"><div class="star ' + cls + '"></div><span class="label">' + filter.name + '</span></li>';
    }).bind(this));

    this.filters.innerHTML = html;

    this.filters.childNodes.each((function (n) {
        if (n.classList.contains('item')) {
            var star = n.childNodes[0];
            var did = n.attributes['data-id'].value;

            var filter = this._filter_map[did];

            if (filter == this._active_filter) {
                this._active_filter_elem = n;
            }

            n.addEventListener('click', (function() {
                this._on_filter_click.bind(this)(n, filter);
            }).bind(this));

            star.addEventListener('mousedown', (function(e) {
                e.preventDefault();
                this._on_filter_star_click.bind(this)(n, star, filter);
            }).bind(this));
        }
    }).bind(this));
}

App.prototype.run = function() {
}

/* vi:ts=4:et */
