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
    this._bugs = [];

    this.searches['search-filters'].on_update = this.on_search_filters.bind(this);

    this.searches['search-bugs'].update_delay = 300;
    this.searches['search-bugs'].on_update = this.on_search_bugs.bind(this);

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

App.prototype.on_search_bugs = function(search) {
    this._update_bugs_list();
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

App.prototype._required_products_from_query = function(query) {
    var products = [];

    // Query for bugs of new products if necessary
    query.products.each(function (p) {
        var product = this.db.product_name_to_product[p];

        if (product) {
            products.push(product);
        }
    });

    query.product_ids.each((function (p) {
        var product = this.db.product_id_to_product[p];

        if (product) {
            products.push(product);
        }
    }).bind(this));

    return products;
}

App.prototype._render_bugs_list = function() {

}

App.prototype._bugs_union = function(a, b) {
    var isa = (typeof a === 'function');
    var isb = (typeof b === 'function');

    if (isa && isb) {
        return function(bug) {
            return a(bug) || b(bug);
        }
    }

    if (!isa && !isb) {
        // Simple union of both
        var ret = {};

        for (var k in a) {
            ret[k] = a[k];
        }

        for (var k in b) {
            if (!(k in ret)) {
                ret[k] = b[k];
            }
        }

        return ret;
    } else {
        if (isa) {
            return b;
        } else {
            return a;
        }
    }
}

App.prototype._bugs_filter = function(a, b) {
    var ret = {};

    for (var k in a) {
        var bug = a[k];

        if (b(bug)) {
            ret[k] = bug;
        }
    }

    return ret;
}

App.prototype._bugs_intersect = function(a, b) {
    var isa = (typeof a === 'function');
    var isb = (typeof b === 'function');

    if (isa && isb) {
        return function(bug) {
            return a(bug) && b(bug);
        }
    }

    if (!isa && !isb) {
        // Simple intersection of both
        var ret = {};

        for (var k in a) {
            if (k in b) {
                ret[k] = a[k]
            }
        }

        return ret;
    } else if (isa) {
        // filter b by a
        return this._bugs_filter(b, a);
    } else {
        // filter a by b
        return this._bugs_filter(a, b);
    }
}

App.prototype._do_query_node = function(node, cb) {
    if (typeof node === 'string') {
        // Simply do a filter
        cb(function(bug) {
            return bug.summary.indexOf(node) != -1;
        });
    } else if ('field' in node) {
        if (node.field == 'product' || node.field == 'product-id') {
            var product;

            if (node.field == 'product') {
                product = this.db.product_name_to_product[node.value];
            } else {
                product = this.db.product_id_to_product[node.value];
            }

            if (product) {
                this.db.bugs().index('product').only(product.name).all((function (bugs) {
                    var ret = {};

                    bugs.each(function(b) {
                        ret[b.id] = b;
                    });

                    cb(ret);
                }).bind(this));
            } else {
                cb({});
            }
        } else {
            cb(function(bug) {
                return bug[node.field].indexOf(node.value) != -1;
            });
        }
    } else {
        if (node.left && node.right) {
            this._do_query_node(node.left, (function(retleft) {
                this._do_query_node(node.right, (function(retright) {
                    if (node.type == 'either') {
                        cb(this._bugs_union(retleft, retright));
                    } else {
                        cb(this._bugs_intersect(retleft, retright));
                    }
                }).bind(this));
            }).bind(this));
        } else if (node.left) {
            this._do_query_node(node.left, cb);
        } else if (node.right) {
            this._do_query_node(node.right, cb);
        } else {
            cb({});
        }
    }
}

App.prototype._update_bugs_list_with_query = function(query) {
    // Perform actual query against database
    this._do_query_node(query.tree, (function(ret) {
        this._bugs = [];

        for (k in ret) {
            this._bugs.push(ret[k]);
        }

        this._render_bugs_list();
    }).bind(this));
}

App.prototype._update_bugs_list_with_filters = function(filters) {
    var q = '!(' + filters.map(function(e) { return e.query; }).join(' ') + ')';
    var s = this.searches['search-bugs'].value();

    if (s) {
        q += ' (' + s + ')';
    }

    var q = new Query(q);
    var products = this._required_products_from_query(q);

    if (products.length == 0) {
        this._bugs = [];
        this._render_bugs_list();
        return;
    }

    var nensure = products.length;

    products.each((function(product) {
        this.db.ensure_product(product.id, (function() {
            nensure--;

            if (nensure == 0) {
                this._update_bugs_list_with_query(q);
            }
        }).bind(this));
    }).bind(this));
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
