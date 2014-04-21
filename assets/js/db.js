var DB = function() {
    //var req = indexedDB.deleteDatabase('bugzini');
    //req.onsuccess = this.open.bind(this);
    this.open();

    this._needs_init_filters = false;
    this.on_filters_updated = function () {};
    this.db = null;
    this.loaded = null;

    this.products = [];
    this.product_id_to_product = {};
    this.product_name_to_product = {};

    return this;
}

DB.prototype.Store = function(db, name) {
    this._db = db;
    this._name = name;
    this._index = null;
    this._range = null;
    this._reverse = false;

    this.sort = null;
    this.read = null;

    return this;
}

DB.prototype.Store.prototype.index = function(name) {
    var ret = this._copy();
    ret._index = name;

    return ret;
}

DB.prototype.Store.prototype.only = function(val) {
    var ret = this._copy();

    ret._range = IDBKeyRange.only(val);
    return ret;
}

DB.prototype.Store.prototype.upper = function(val) {
    var ret = this._copy();

    ret._range = IDBKeyRange.upperBound(val);
    return ret;
}

DB.prototype.Store.prototype.lower = function(val) {
    var ret = this._copy();

    ret._range = IDBKeyRange.lowerBound(val);
    return ret;
}
DB.prototype.Store.prototype.bound = function(lower, upper) {
    var ret = this._copy();

    ret._range = IDBKeyRange.bound(lower, upper);
    return ret;
}

DB.prototype.Store.prototype._copy = function() {
    var ret = new this._db.Store(this._db, this._name);

    ret._index = this._index;
    ret._range = this._range;
    ret._reverse = this._reverse;

    ret.sort = this.sort;
    ret.read = this.read;

    return ret;
}

DB.prototype.Store.prototype.put = function(item, cb) {
    var tr = this._db.db.transaction(this._name, 'readwrite');
    var store = tr.objectStore(this._name);

    if (cb) {
        tr.oncomplete = (function() {
            cb();
        }).bind(this);
    }

    store.put(item);
}

DB.prototype.Store.prototype.find = function(key, cb) {
    var tr = this._db.db.transaction(this._name);
    var store = tr.objectStore(this._name);

    var req = store.openCursor(IDBKeyRange.only(key));

    req.onsuccess = function(e) {
        var cursor = e.target.result;

        if (cursor) {
            cb(cursor.value);
        } else {
            cb(null);
        }
    }
}

DB.prototype.Store.prototype.reverse = function() {
    var ret = this._copy();
    ret._reverse = !this._reverse;
    return ret;
}

DB.prototype.Store.prototype.cursor = function(cb, oncomplete) {
    var tr = this._db.db.transaction(this._name);
    var ret = [];

    tr.oncomplete = oncomplete;
    var store = tr.objectStore(this._name);

    if (this._index) {
        store = store.index(this._index);
    }

    var req;

    if (this._range) {
        if (this._reverse) {
            req = store.openCursor(this._range, 'prev');
        } else {
            req = store.openCursor(this._range);
        }
    } else {
        if (this._reverse) {
            req = store.openCursor(null, 'prev');
        } else {
            req = store.openCursor();
        }
    }

    req.onsuccess = function(e) {
        cb(e.target.result);
    }
}

DB.prototype.Store.prototype.all = function(cb) {
    var ret = [];

    var oncomplete = (function() {
        if (this.sort != null)
        {
            ret = ret.sort(this.sort);
        }

        cb(ret);
    }).bind(this);

    this.cursor((function(cursor) {
        if (!cursor) {
            return;
        }

        var record = cursor.value;

        if (this.read != null)
        {
            this.read(record);
        }

        ret.push(record);
        cursor.continue();
    }).bind(this), oncomplete);
}

DB.prototype.filters = function() {
    var ret = new this.Store(this, 'filters');

    ret.sort = function(a, b) {
        if (a.starred != b.starred) {
            return a.starred ? -1 : 1;
        }

        if (a.is_product != b.is_product) {
            return b.is_product ? -1 : 1;
        }

        var n1 = a.name_case;
        var n2 = b.name_case;

        return n1.localeCompare(n2);
    };

    ret.read = function(a) {
        a.name_case = a.name.toLowerCase();
    }

    return ret;
}

DB.prototype.bugs = function() {
    var ret = new this.Store(this, 'bugs');

    return ret;
}

DB.prototype.open = function () {
    var req = indexedDB.open('bugzini', 1);

    req.onsuccess = this.open_success.bind(this);
    req.onerror = this.open_error.bind(this);
    req.onupgradeneeded = this.open_upgrade_needed.bind(this);

    return this;
}

DB.prototype.open_success = function(e) {
    this.db = e.target.result;

    this.init_filters((function() {
        if (this.loaded) {
            this.loaded();
        }
    }).bind(this));
}

DB.prototype.open_error = function(e) {
}

DB.prototype.open_upgrade_needed = function(e) {
    this.db = e.target.result;

    switch (e.newVersion)
    {
    case 1:
        this.upgrade_v1();
        break;
    }
}

DB.prototype.upgrade_v1 = function() {
    var filters = this.db.createObjectStore('filters', { keyPath: 'id' });
    filters.createIndex('starred', 'starred', { unique: false });
    filters.createIndex('is_product', 'is_product', { unique: false });
    filters.createIndex('name', 'name', { unique: false });

    this.db.createObjectStore('products', { keyPath: 'id' });

    var bugs = this.db.createObjectStore('bugs', { keyPath: 'id' });
    bugs.createIndex('starred', 'starred', { unique: false });
    bugs.createIndex('is_unread', 'is_unread', { unique: false });
    bugs.createIndex('product', 'product', { unique: false });
    bugs.createIndex('product_open', ['product', 'is_open'], { unique: false });
    bugs.createIndex('last_change_time', 'last_change_time', { unique: false });
    bugs.createIndex('product_last_change_time', ['product', 'last_change_time'], { unique: false });

    this._needs_init_filters = true;
}

DB.prototype.init_filters = function(cb) {
    if (this._needs_init_filters) {
        this.init_filters_load();
        cb();
    } else {
        this.filters().all((function(filters) {
            filters.each((function (filter) {
                if (filter.is_product) {
                    this.products.push(filter);
                    this.product_id_to_product[filter.id] = filter;
                    this.product_name_to_product[filter.name.toLowerCase()] = filter;
                }
            }).bind(this));

            cb();
            this.on_filters_updated();
        }).bind(this));
    }
}

DB.prototype.init_filters_load = function() {
    Service.get('/product/all', {
        success: (function(req, ret) {
            // Load products into filters
            var tr = this.db.transaction('filters', 'readwrite');
            var store = tr.objectStore('filters');

            this.products = ret;
            this.product_id_to_product = {};
            this.product_name_to_product = {};

            tr.oncomplete = (function(e) {
                this.on_filters_updated();
            }).bind(this);

            ret.each((function (product) {
                this.product_id_to_product[product.id] = product;
                this.product_name_to_product[product.name.toLowerCase()] = product;

                var filter = {
                    description: product.description,
                    name: product.name,
                    id: product.id,
                    query: 'product-id:' + product.id,
                    color: '#268BD2',
                    is_product: true
                }

                store.put(filter);
            }).bind(this));
        }).bind(this),

        error: (function(req) {
            console.log([req.status, req.statusText]);
        }).bind(this)
    });
}

DB.prototype._store_bugs = function(bugs, fixit, cb) {
    var tr = this.db.transaction('bugs', 'readwrite');
    var store = tr.objectStore('bugs');

    tr.oncomplete = function() {
        if (cb) {
            cb();
        }
    }

    bugs.each((function (bug) {
        if (fixit) {
            bug.is_open = bug.is_open ? 1 : 0;
            bug.creation_time = Date.parse(bug.creation_time);
            bug.last_change_time = Date.parse(bug.last_change_time);
            bug._component_ci = bug.component.toLowerCase();
            bug._severity_ci = bug.severity.toLowerCase();
            bug._component_ci = bug.component.toLowerCase();
            bug._status_ci = bug.status.toLowerCase();
        }

        store.put(bug);
    }).bind(this));
}

DB.prototype._process_bugs = function(product, bugs, cb) {
    var tr = this.db.transaction('products', 'readwrite');
    var store = tr.objectStore('products');

    store.put({id: product, last_update: new Date()});

    // Update bugs
    this._store_bugs(bugs, true, cb);
}

DB.prototype.ensure_product = function(id, cb) {
    var store = new this.Store(this, 'products');

    store.find(id, (function(record) {
        if (record) {
            cb();
        } else {
            Service.get('/product/' + id + '/bugs', {
                success: (function(req, ret) {
                    this._process_bugs(id, ret, cb);
                }).bind(this)
            });
        }
    }).bind(this));
}

DB.prototype._ensure_comments = function(bug, cb) {
    var after;

    if (bug.comments && bug.comments.length > 0) {
        after = (bug.comments[bug.comments.length - 1].time / 1000);
    } else {
        after = 0;
    }

    Service.get('/bug/' + bug.id + '/comments?after=' + after, {
        success: (function(req, ret) {
            for (var i = 0; i < ret.length; i++) {
                ret[i].is_unread = 1;
                ret[i].time = Date.parse(ret[i].time);
            }

            if (!bug.comments) {
                bug.comments = ret;
            } else {
                bug.comments = bug.comments.concat(ret);
            }

            this._store_bugs([bug], false, function() {
                this._mark_read(bug, cb(bug));
            });
        }).bind(this)
    })
}

DB.prototype._mark_read = function(bug, val) {
    if (!val) {
        return;
    }

    var needssave = false;

    if (bug.is_unread) {
        needssave = true;
    }

    bug.is_unread = 0;

    if (bug.comments) {
        for (var i = 0; i < bug.comments.length; i++) {
            if (bug.comments[i].is_unread) {
                needssave = true;
            }

            bug.comments[i].is_unread = 0;
        }
    }

    if (needssave) {
        this._store_bugs([bug], false);
    }
}

DB.prototype.ensure_bug = function(id, cb) {
    var store = this.bugs();

    store.find(id, (function(record) {
        if (record && record.comments && record.comments.length > 0) {
            if (record.is_unread) {
                cb(record, true);
                this._ensure_comments(record, cb);
            } else {
                this._mark_read(record, cb(record));
            }
        } else if (!record) {
            Service.get('/bug/' + id, {
                success: (function(req, ret) {
                    if (ret.comments) {
                        for (var i = 0; i < ret.comments.length; i++) {
                            ret.comments[i].time = Date.parse(ret.comments[i].time);
                        }
                    }

                    this._ensure_comments(ret, cb);
                }).bind(this)
            });
        } else {
            cb(record, true);
            this._ensure_comments(record, cb);
        }
    }).bind(this));
}

DB.prototype.update_product = function(id, cb) {
    this.filters().find(id, (function(record) {
        if (!record) {
            cb();
            return;
        }

        var name = record.name;
        var store = this.bugs().index('product_last_change_time').reverse().bound([name], [name + '\t']);

        store.cursor((function(cursor) {
            var bug = cursor.value;

            Service.get('/product/' + id + '/bugs?after=' + bug.last_change_time / 1000, {
                success: (function(req, ret) {
                    for (var i = 0; i < ret.length; i++) {
                        ret[i].is_unread = 1;
                    }

                    this._process_bugs(id, ret, cb);
                }).bind(this)
            });
        }).bind(this));
    }).bind(this));
}

/* vi:ts=4:et */
