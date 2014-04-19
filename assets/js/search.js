var Search = function(elem) {
    this.elem = elem;

    var e = new $(elem);

    this._input = e.query('input');
    this._close = e.query('.close-icon');
    this._params = e.query('.params-icon');

    this._all = [this._input, elem, this._close, this._params];

    this._input.addEventListener('focus', this._on_focus.bind(this));
    this._input.addEventListener('blur', this._on_blur.bind(this));
    this._input.addEventListener('keydown', this._on_keydown.bind(this));
    this._input.addEventListener('change', this._on_change.bind(this));
    this._input.addEventListener('input', this._on_change.bind(this));
    this._input.addEventListener('paste', this._on_change.bind(this));
    this._close.addEventListener('click', this._on_close.bind(this));

    this._last_value = '';
    this._on_update_timeout = 0;
    this.update_delay = 0;

    this.on_update = function() {}
}

Search.prototype.value = function() {
    return this._input.value;
}

Search.prototype._on_focus = function() {
    this._all.each(function (item) {
        item.classList.add('focus');
    });
}

Search.prototype._on_blur = function() {
    this._all.each(function (item) {
        item.classList.remove('focus');
    });
}

Search.prototype._on_close = function() {
    this._cancel();
}

Search.prototype._emity_update = function(force) {
    if (this._on_update_timeout != 0) {
        clearTimeout(this._on_update_timeout);
        this._on_update_timeout = 0;
    }

    if (force || this.update_delay == 0) {
        this.on_update(this);
    } else {
        this._on_update_timeout = setTimeout((function() {
            this._on_update_timeout = 0;
            this.on_update(this);
        }).bind(this), this.update_delay);
    }
}

Search.prototype._on_change = function() {
    var val = this.value();

    if (val != this._last_value) {
        this._last_value = val;
        this._emity_update(false);
    }
}

Search.prototype._cancel = function() {
    this._input.value = '';
    this._emity_update(true);

    this._input.focus();
}

Search.prototype._on_keydown = function(e) {
    if (e.which == 27) {
        this._cancel();
    } else {
        this._on_change();
    }
}

/* vi:ts=4:et */
