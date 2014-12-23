var Query = function(q) {
    this.query = q;
    this.open_only = true;

    this.tree = this._parse_query(q);

    this._extract_products();
    return this;
}

Query.prototype._tokenize_query = function(query) {
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

            var fieldop = [':', '>', '<'];
            var val = query.slice(s, i);
            var tp = 's';

            for (var k = 0; k < fieldop.length; k++) {
                var pos = val.indexOf(fieldop[k]);

                if (pos != -1) {
                    val = {
                        op: fieldop[k],
                        field: val.slice(0, pos),
                        value: val.slice(pos + 1, val.length)
                    }

                    if (val.field === 'status' && val.value.toLowerCase() === 'resolved')
                    {
                        this.open_only = false;
                    }

                    tp = 'f';
                    break;
                }
            }

            tokens.push({type: tp, value: val});
            i--;
            break;
        }

        i++;
    }

    return tokens;
}

Query.prototype._parse_tokens = function(tokens) {
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
        case 'f':
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

Query.prototype._binary_query = function(query) {
    if (typeof query === 'string') {
        return query;
    }

    if (!query.hasOwnProperty('either')) {
        return query;
    }

    var node = null;

    if (query.either.length != 0) {
        node = this._binary_query(query.either[0]);

        for (var i = 1; i < query.either.length; i++) {
            node = {type: 'either', left: node, right: this._binary_query(query.either[i])};
        }
    }

    if (query.all.length != 0) {
        if (node == null) {
            node = this._binary_query(query.all[0]);
        } else {
            node = {type: 'all', left: node, right: this._binary_query(query.all[0])};
        }

        for (var i = 1; i < query.all.length; i++) {
            node = {type: 'all', left: node, right: this._binary_query(query.all[i])};
        }
    }

    return node;
}

Query.prototype._parse_query = function(query) {
    var tokens = this._tokenize_query(query);
    var q = this._parse_tokens(tokens);

    return this._binary_query(q);
}

Query.prototype._extract_products = function() {
    var s = [this.tree];
    this.products = [];
    this.product_ids = [];

    while (s.length > 0) {
        var item = s.shift();

        if (!item) {
            continue;
        }

        if (item.hasOwnProperty('left') && item.hasOwnProperty('right')) {
            s.push(item.left);
            s.push(item.right);
        } else if (item.hasOwnProperty('field') && item.hasOwnProperty('value')) {
            if (item.field == 'product') {
                this.products.push(item.value);
            } else if (item.field == 'product-id') {
                this.product_ids.push(parseInt(item.value));
            }
        }
    }
}
